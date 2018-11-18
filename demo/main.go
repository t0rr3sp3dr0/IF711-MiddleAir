//go:generate protoc -I . --go_out=. server.proto

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/t0rr3sp3dr0/middleair/client"
	"github.com/t0rr3sp3dr0/middleair/server"
	"github.com/t0rr3sp3dr0/middleair/util"
)

func main() {
	go func() {
		defer func() {
			if errs := client.ClosePersistentConns(); errs != nil {
				panic(errs)
			}
		}()

		scanner := bufio.NewScanner(os.Stdin)

	main:
		for {
			fmt.Println("MiddleAir Demo Client")
			fmt.Println()
			fmt.Println("1 - Remote Shell")
			fmt.Println("2 - Text to Speech")
			fmt.Println("0 - Quit")

		mainScan:
			for {
				fmt.Println()
				fmt.Print(">: ")

				scanner.Scan()
				i, err := strconv.Atoi(scanner.Text())
				if err != nil {
					log.Println(err)
					continue mainScan
				}

				switch i {
				case 1:
					req := &RemoteShellRequest{}
					res := &RemoteShellResponse{}
					opt := &client.Options{}

					fmt.Println()

					fmt.Print(">: [name] ")
					scanner.Scan()
					req.Name = scanner.Text()

					fmt.Print(">: [#args] ")
					scanner.Scan()
					nArgs, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Println(err)
						continue mainScan
					}
					for i := 0; i < nArgs; i++ {
						fmt.Printf(">: [agr#%d] ", i)
						scanner.Scan()
						req.Args = append(req.Args, scanner.Text())
					}

					fmt.Print(">: [stdin] ")
					scanner.Scan()
					req.Stdin = []byte(scanner.Text())

					log.Println(req, "\n")

					fmt.Print(">: [#tags] ")
					scanner.Scan()
					nTags, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Println(err)
						continue mainScan
					}
					for i := 0; i < nTags; i++ {
						fmt.Printf(">: [tag#%d] ", i)
						scanner.Scan()
						opt.Tags = append(opt.Tags, scanner.Text())
					}

					fmt.Print(">: [strictMatch] ")
					scanner.Scan()
					opt.StrictMatch = scanner.Text() != "0"

					fmt.Print(">: [broadcast] ")
					scanner.Scan()
					opt.Broadcast = scanner.Text() != "0"

					fmt.Print(">: [persistent] ")
					scanner.Scan()
					opt.Persistent = scanner.Text() != "0"

					fmt.Print(">: [#credentials] ")
					scanner.Scan()
					nCredentials, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Println(err)
						continue mainScan
					}
					for i := 0; i < nCredentials; i++ {
						fmt.Printf(">: [credential#%d] ", i)
						scanner.Scan()
						b, err := strconv.Atoi(scanner.Text())
						if err != nil {
							log.Println(err)
							continue mainScan
						}
						opt.Credentials = append(opt.Credentials, byte(b))
					}

					log.Println(opt, "\n")

					if err := client.Invoke(req, res, opt); err != nil {
						log.Println(err)
						continue mainScan
					}
					log.Println(res, "\n")

					break mainScan

				case 2:
					req := &TextToSpeechRequest{}
					res := &TextToSpeechResponse{}
					opt := &client.Options{}

					fmt.Println()

					fmt.Print(">: [message] ")
					scanner.Scan()
					req.Message = scanner.Text()

					log.Println(req, "\n")

					fmt.Print(">: [#tags] ")
					scanner.Scan()
					nTags, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Println(err)
						continue mainScan
					}
					for i := 0; i < nTags; i++ {
						fmt.Printf(">: [tag#%d] ", i)
						scanner.Scan()
						opt.Tags = append(opt.Tags, scanner.Text())
					}

					fmt.Print(">: [strictMatch] ")
					scanner.Scan()
					opt.StrictMatch = scanner.Text() != "0"

					fmt.Print(">: [broadcast] ")
					scanner.Scan()
					opt.Broadcast = scanner.Text() != "0"

					fmt.Print(">: [persistent] ")
					scanner.Scan()
					opt.Persistent = scanner.Text() != "0"

					fmt.Print(">: [#credentials] ")
					scanner.Scan()
					nCredentials, err := strconv.Atoi(scanner.Text())
					if err != nil {
						log.Println(err)
						continue mainScan
					}
					for i := 0; i < nCredentials; i++ {
						fmt.Printf(">: [credential#%d] ", i)
						scanner.Scan()
						b, err := strconv.Atoi(scanner.Text())
						if err != nil {
							log.Println(err)
							continue mainScan
						}
						opt.Credentials = append(opt.Credentials, byte(b))
					}

					log.Println(opt, "\n")

					if err := client.Invoke(req, res, opt); err != nil {
						log.Println(err)
						continue mainScan
					}
					log.Println(res, "\n")

					break mainScan

				case 0:
					os.Exit(0)
					break main

				default:
					log.Println(errors.New("Invalid Option"))
				}
			}
		}
	}()

	opt := util.Options{
		Port:     1337,
		Protocol: "tcp",
	}

	for {
		invoker, err := server.NewInvoker(&Server{}, opt)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := invoker.Accept([]byte{0, 1, 2, 3}); err != nil {
			log.Println(err)
			continue
		}

		go func() {
			if err := invoker.Loop(); err != nil {
				panic(err)
			}
		}()
	}
}
