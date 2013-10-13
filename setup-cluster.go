package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"github.com/howeyc/gopass"
	"code.google.com/p/go.crypto/ssh"
)

type password string
func (p password) Password(_ string) (string, error) {
    return string(p), nil
}

var keyfile_path = "~/.ssh/id_rsa.pub"

func Exists(name string) bool {
    if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
    return true
}


func scp_key_to(host string) {
	//if NO keyfile exists
	if !Exists(keyfile_path) {
		fmt.Printf("   no such file or directory: %s\n", keyfile_path)
		out, err := exec.Command( //generate keyfile
			"sh", "-c",
			fmt.Sprintf("ssh-keygen -q -t rsa -f %s -N \"\"", keyfile_path)).Output()
		if err != nil {
			fmt.Printf("%s\n", out)
			log.Fatal(err)
		}
	}
	out, err := exec.Command( //scp keyfile to host
		fmt.Sprintf("scp %s %s:~/.ssh/id_rsa.pub", keyfile_path, host)).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
}


func ssh_is_avail_on(host string) (is_available bool) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:22", host))
	if err != nil {
		log.Print(err)
		return false
	}
	conn.Close()
	return true
}


func _exec_through_ssh(command string, host string, user string, pass []byte) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(password(pass)),
		},
	}
	client, err := ssh.Dial("tcp",fmt.Sprintf("%s:22",host), config)
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
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		panic("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}


func add_key_to_authorized_on(host string) {
	command := "cat id_rsa.pub >> authorized_keys"
	u, _ := user.Current()
	user := u.Username
	fmt.Printf("enter assword for %s@%s: ", user, host)
    pass := gopass.GetPasswd()
	_exec_through_ssh(command, host, user, pass)
}


func ConfigSshOn(host string) {
	fmt.Printf(" - %s ...\n", host)

	if ssh_is_avail_on(host) {
		scp_key_to(host)
		add_key_to_authorized_on(host)
	} else { 
		fmt.Printf("  No ssh available at %s, skiping\n", host)
	}
	fmt.Printf("Done\n\n")
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
