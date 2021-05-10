package handlers

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/NHAS/reverse_ssh/internal"
	"github.com/NHAS/reverse_ssh/internal/server/terminal"
	"github.com/NHAS/reverse_ssh/internal/server/terminal/commands"
	"github.com/NHAS/reverse_ssh/internal/server/terminal/commands/constants"
	"github.com/NHAS/reverse_ssh/internal/server/users"
	"github.com/NHAS/reverse_ssh/pkg/trie"
	"golang.org/x/crypto/ssh"
)

func Shell(controllableClients *sync.Map, autoCompleteClients *trie.Trie) internal.ChannelHandler {

	return func(user *users.User, newChannel ssh.NewChannel) {

		defer log.Printf("Human client disconnected %s (%s)\n", user.ServerConnection.RemoteAddr(), user.ServerConnection.ClientVersion())

		// At this point, we have the opportunity to reject the client's
		// request for another logical connection
		connection, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel (%s)", err)
			return
		}
		defer connection.Close()

		user.ShellConnection = connection
		user.ShellRequests = requests

		term := terminal.NewAdvancedTerminal(connection, "> ")

		term.AddValueAutoComplete(constants.RemoteId, autoCompleteClients)

		defaultHandle := internal.NewDefaultHandler(user, term)

		term.AddCommand("ls", commands.List(controllableClients))
		term.AddCommand("help", commands.Help())
		term.AddCommand("exit", commands.Exit())
		term.AddCommand("connect", commands.Connect(user, defaultHandle, controllableClients))
		term.AddCommand("rc", commands.RC(user, controllableClients))
		term.AddCommand("proxy", commands.Proxy(user, controllableClients))

		// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
		// While we arent passing the requests directly to the remote host consume them with our terminal and store the results to send initialy to the remote on client connect
		defaultHandle.Start()

		//Send list of controllable remote hosts to human client
		fmt.Fprintf(term, "Connected controllable clients: \n")
		controllableClients.Range(func(idStr interface{}, value interface{}) bool {
			fmt.Fprintf(term, "%s, client version: %s\n",
				idStr,
				value.(ssh.Conn).ClientVersion(),
			)
			return true
		})

		//Blocking function to handle all the human function calls. Will return io.EOF on exit, otherwise an error is passed up we cant deal with
		err = term.Run()
		if err != nil && err != io.EOF {
			fmt.Fprintf(term, "Error: %s\n", err)
		}

	}

}
