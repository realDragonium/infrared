package command

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"runtime"
	"strings"
)

func Read(r io.Reader) {
	if r == nil {
		r = os.Stdin
	}

	reader := bufio.NewReader(r)

	for {
		commandString, err := reader.ReadString('\n')
		if err != nil {
			log.Err(err).Msg("Error while parsing commandString")
			continue
		}

		switch runtime.GOOS {
		case "windows":
			commandString = strings.Replace(commandString, "\r\n", "", -1)
		default:
			commandString = strings.Replace(commandString, "\n", "", -1)
		}

		/*found := false
		for command, params := range commands {
			if strings.HasPrefix(command, commandString) {
				continue
			}
			found = true
		}

		fmt.Fscanf()*/
	}
}
