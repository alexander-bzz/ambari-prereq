package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"flag"
	"fmt"
	"github.com/howeyc/gopass"
	"log"
	"net"
	"os"
	"os/exec"
)

type password string

func (p password) Password(_ string) (string, error) {
	return string(p), nil
}

func CurrentUserPath() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		log.Fatal(err)
	}
	s := string(bytes.TrimSpace(out))
	fmt.Printf("\tUser name is: '%s'\n", s)
	return s
}

var keyfile_path = "./.ssh/id_rsa.pub"

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

func scp_key_to(host string) {
	//if NO keyfile exists
	if !Exists(keyfile_path) {
		fmt.Printf("\tNo such file or directory: %s\n   So generateing a new key", keyfile_path)
		out, err := exec.Command( //generate keyfile
			"sh", "-c",
			fmt.Sprintf("ssh-keygen -q -t rsa -f %s -N \"\"", keyfile_path)).Output()
		if err != nil {
			fmt.Printf("%s\n", out)
			log.Fatal(err)
		}
	} else {
		fmt.Printf("\tUsing a keyfile: %s\n", keyfile_path)
	}
	//scp keyfile to host
	fmt.Printf("\tCoping a keyfile: %s to %s using scp ...\n", keyfile_path, host)

	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("scp %s %s:~/.ssh/id_rsa.pub", keyfile_path, host))
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal("\tFailed to copy keyfile" + err.Error())
	}
	fmt.Printf("\tDone\n%s\n", out)
}

func _exec_through_ssh(command string, host string, user string, pass []byte) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(password(pass)),
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
	//var b bytes.Buffer
	//session.Stdout = &b
	b, err := session.CombinedOutput(command)
	if err != nil {
		fmt.Println(string(b))
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(string(b))
}

func add_key_to_authorized_on(host string) {
	fmt.Printf("\tAdding a new key to autorized_keys ...\n")
	command := "cat ./.ssh/id_rsa.pub >> ./.ssh/authorized_keys"
	user :=  CurrentUserPath()
	fmt.Printf("\t\tEnter password for %s@%s: ", user, host)
	pass := gopass.GetPasswd()
	_exec_through_ssh(command, host, user, pass)
	fmt.Printf("\tDone\n")
}



func ConfigSshOn(host string) {
	fmt.Printf(" - %s ...\n", host)

	if ssh_is_avail_on(host) {
		scp_key_to(host)
		add_key_to_authorized_on(host)
		//TODO(alex) add erpl repository
		//wget http://public-repo-1.hortonworks.com/ambari/centos6/1.x/GA/ambari.repo
		//cp ambari.repo /etc/yum.repos.d
		//TODO(alex) install ambari
		//yum install epel-release
		//yum install ambari-server
	} else {
		fmt.Printf("  No ssh available at %s, skiping\n", host)
	}
	fmt.Printf("   Done\n\n")
}

func main() {
	flag.Parse()
	hosts := flag.Args()

	fmt.Printf("Pre-configuring hosts for Ambari cluster:\n")
	for _, host := range hosts {
		ConfigSshOn(host)

	}
	/*read list of hosts
	  for each Host
	    scp ~/.ssh/id_rsa.pub Host:~/.ssh/id_rsa.pub
	    ssh Host <enter password>
	      cat id_rsa.pub >> authorized_keys
	      chmod 700 ~/.ssh
	      chmod 600 ~/.ssh /authorized_keys

	*/
}
