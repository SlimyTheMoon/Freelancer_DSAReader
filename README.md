# Freelancer_DSAReader

Using this tool is straight forward:

First, for linux and windows put the corresponding file in a folder where the programm shall run.


Under Windows, run the DSAReader.exe,
after that, if you are asked if you want this application to be available publicly or in lan, feel free to cancel, for me it worked then on my pc anyway, then navigate to https://localhost:8443 in your browser.

Under linux, you first need to make it executable using:

```
chmod +x DSAReader-linux
```
and then you can run it using:

```
./DSAReader-linux
```

directly from the terminal.


In case you want to compile the source code on your own:

Go to the official website of go,
you can google it or take this link:
https://go.dev/

Install it for your operating system, so we can proceed.

UNDER WINDOWS:

First, put the main.go you can download under releases or from the repository itself in a folder of your choice, maybe even, where the programm shall run.

Then you open your terminal at this place and run the following command:

```
go build -o DSAReader.exe main.go
```

Now the executable is ready.


UNDER LINUX:

First, put the main.go you can download under releases or from the repository itself in a folder of your choice, maybe even, where the programm shall run.

then run the following command from your terminal from the directory the main.go resides in, make sure, the architecture matches the one you are using:

```
GOOS=linux GOARCH=amd64 go build -o freelancer-reader.exe main.go
```

Now you can proceed with the how to use above!