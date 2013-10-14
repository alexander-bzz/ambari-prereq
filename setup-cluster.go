package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"flag"
	"fmt"
	"github.com/howeyc/gopass"
	"log"
	"net"
	"io"
	"os"
	"os/exec"
)

type password string

func (p password) Password(_ string) (string, error) {
	return string(p), nil
}

func CurrentUserName() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		log.Fatal(err)
	}
	s := string(bytes.TrimSpace(out))
	fmt.Printf("Using current user name: '%s'\n", s)
	return s
}

var private_keyfile_path = "./.ssh/id_rsa"
var public_keyfile_path = private_keyfile_path + ".pub"


func Exists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func ssh_is_avail_on(host string) (is_available bool) {
	fmt.Printf("\tChecking SSH at %s:22 ...", host)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:22", host))
	if err != nil {
		fmt.Printf("Fail\n")
		log.Print(err)
		return false
	}
	conn.Close()
	fmt.Printf("OK\n")
	return true
}

func scp_key_to(host string, username *string, password *string) {
	//if NO keyfile exists
	if !Exists(public_keyfile_path) {
		fmt.Printf("\tNo such file or directory: %s\n\tSo generateing a new key ...", public_keyfile_path)
		out, err := exec.Command( //generate keyfile
			"sh", "-c",
			fmt.Sprintf("ssh-keygen -q -t rsa -f %s -N \"\"", private_keyfile_path)).Output()
		if err != nil {
			fmt.Printf("%s\n", out)
			log.Fatal(err)
		}
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("\tUsing a keyfile: %s\n", public_keyfile_path)
	}
	//scp keyfile to host
	fmt.Printf("\tCoping a keyfile: %s to %s using scp ...\n", public_keyfile_path, host)

	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("scp %s %s:~/.ssh/id_rsa.pub", public_keyfile_path, host))
	//cmd.Stdin = os.Stdin
	stdin, err := cmd.StdinPipe()
    if err != nil {
        log.Panic(err)
    }
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("\tFailed to copy keyfile" + err.Error())
	}
	fmt.Printf("\tDone\n%s\n", out)
	if password != nil {
		io.Copy(stdin, bytes.NewBufferString(*password+"\n"))
	}
}

func _exec_through_ssh(command string, host string, user *string, pass *string) {
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(password(*pass)),
		},
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		log.Fatal("Failed to dial: " + err.Error())
		return
	}
	defer client.Close()
	//
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: " + err.Error())
		return
	}
	defer session.Close()
	//now we can run commands!
	b, err := session.CombinedOutput(command)
	if err != nil {
		fmt.Println(string(b))
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(string(b))
}

func add_key_to_authorized_on(host string, username *string, password *string) {
	fmt.Printf("\tAdding a new key to autorized_keys ...\n")
	command := "cat ./.ssh/id_rsa.pub >> ./.ssh/authorized_keys"
	if password == nil {
		fmt.Printf("\t\tEnter password for %s@%s: ", username, host)
		p := string(gopass.GetPasswd())
		password = &p
	}
	_exec_through_ssh(command, host, username, password)
	fmt.Printf("\tDone\n")
}



func ConfigSshOn(host string, username *string, password *string) {
	fmt.Printf(" - %s ...\n", host)

	if ssh_is_avail_on(host) {
		scp_key_to(host, username, password)
		add_key_to_authorized_on(host, username, password)
	} else {
		fmt.Printf("  No ssh available at %s, skiping\n", host)
	}
	fmt.Printf("   Done\n\n")
}

func Exec(comm string) {
	cmd := exec.Command("sh", "-c", comm)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatal("\tFailed to run command " + comm + "\n" + err.Error())
	}
	fmt.Println(string(out))
}


func main() {
	username := flag.String("u", CurrentUserName(), "Username for SSH user")
	password := flag.String("p", *username, "Password for SSH user")
	flag.Parse()

	hosts := flag.Args()
	if len(hosts) == 0 {
		fmt.Printf("Usage:\n ./ambari-prereq [-p <password>] [-u <username>] <hosname1> ... <hostnameN>\n")
		flag.PrintDefaults()
		log.Fatal()
	}

	fmt.Printf("Pre-configuring hosts for Ambari cluster:\n")
	for _, host := range hosts {
		ConfigSshOn(host, username, password)
	}

	fmt.Printf("Adding ambari yum repository...")
	Exec("wget http://public-repo-1.hortonworks.com/ambari/centos6/1.x/GA/ambari.repo")
	Exec("sudo cp ambari.repo /etc/yum.repos.d")
	Exec("sudo yum install epel-release")
	fmt.Printf("Done")


	//TODO(alex) install ambari
	fmt.Printf("Installing ambari-server...")
	Exec("sudo yum install ambari-server")
	fmt.Printf("Done")

	/*read list of hosts
	  for each Host
	    scp ~/.ssh/id_rsa.pub Host:~/.ssh/id_rsa.pub
	    ssh Host <enter password>
	      cat id_rsa.pub >> authorized_keys
	      chmod 700 ~/.ssh
	      chmod 600 ~/.ssh /authorized_keys

	*/
}
