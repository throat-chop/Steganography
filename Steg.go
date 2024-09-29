package main

import (
	"DataGhost/Data"
	"DataGhost/Image"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

func main() {
	args := os.Args

	ctx, cancel := context.WithCancel(context.Background())

	if len(args) == 6 && args[5] == "--safe" {
		args = args[:len(args)-1]
	} else if len(args) == 5 && args[4] == "--safe" {
		args = args[:len(args)-1]
	} else if len(args) == 4 && args[3] == "--safe" {
		args = args[:len(args)-1]
	} else {
		go greeter(ctx)
	}

	if args[1] == "-c" && len(args) < 3 {
		cancel()
		fmt.Print("Usage: ./DataGhost <-c(check capacity)> </path/to/target> <--safe(optional safe mode without graphics)> \n")
		os.Exit(1)

	} else if args[1] != "-c" && (len(args) < 4 || len(args) > 5) {
		cancel()
		fmt.Print("Usage: ./DataGhost <-e/-d/-c(encode/decode/checkcapacity)> </path/to/target> </path/to/data> <passphrase(optional)> <--safe(optional safe mode without graphics)> \n")
		os.Exit(1)

	}

	errc := make(chan error)
	final := make(chan string)
	defer close(errc)
	go work(args, errc, final)

	for {
		select {

		case err := <-errc:
			if err != nil {
				time.Sleep(1 * time.Second)
				cancel()
				time.Sleep(1 * time.Second)
				fmt.Print("\r", "Status: ", err, "\n")
				os.Exit(1)
			}
		case fin := <-final:

			time.Sleep(1 * time.Second)
			cancel()
			time.Sleep(1 * time.Second)
			fmt.Print("\r", "Status: ", fin, " -- Program will exit now\n")
			os.Exit(0)

		default:
			time.Sleep(250 * time.Millisecond)
			fmt.Print("\r", "Status: ", "working..!")
		}
	}

}

func work(args []string, errc chan<- error, fin chan<- string) {
	mode := args[1]
	targetFile := args[2]
	dataFile := ""
	passphrase := ""
	resultfile := path.Dir(targetFile)

	if args[1] != "-c" {
		dataFile = args[3]
		if len(args) == 5 {
			passphrase = args[4]
		}
	}

	if mode == "-e" {
		enc := make(chan *Data.Data)
		defer close(enc)

		go Data.NewData(dataFile, passphrase, errc, enc)
		data := <-enc

		img, err := Image.NewImage(targetFile)
		if err != nil {
			errc <- err
			return
		}
		resultfile = path.Join(resultfile, "hidden.png")

		if err := img.ImgCheck(data.Size); err != nil {
			errc <- err
			return
		}

		err = img.Hide(data.Data, resultfile)
		if err != nil {
			errc <- err
			return
		}
		fin <- "Done! Data saved at: " + resultfile

	} else if mode == "-d" {

		tgt, err := os.Open(targetFile)
		defer tgt.Close()
		if err != nil {
			errc <- err
			return
		}

		img, err := Image.NewImage(targetFile)
		if err != nil {
			errc <- err
			return
		}

		extracted := img.Extract()
		resultfile = path.Join(path.Dir(dataFile), "extracted")

		xtn, err := Data.Decode(resultfile, extracted, passphrase)
		if err != nil {
			errc <- err
			return
		}

		fin <- "Done! Data saved at: " + resultfile + xtn

	} else if mode == "-c" {
		img, err := Image.NewImage(targetFile)
		if err != nil {
			errc <- err
			return
		}
		fin <- "Data hiding capacity of: " + strconv.Itoa(int(img.HidingCapacity)) + " B"
	} else {
		errc <- fmt.Errorf("invalid mode: %s choose '-e','-d' or '-c'", mode)
	}

}

func greeter(ctx context.Context) {
	fmt.Println("Ready to hide?...")

	cmdi := exec.Command("resize", "-s", "40", "128")
	cmd := exec.CommandContext(ctx, "chafa", "greetings.gif")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmdi.Run()
	if err != nil {
	}
	err = cmd.Run()
	if err != nil {
	}
}
