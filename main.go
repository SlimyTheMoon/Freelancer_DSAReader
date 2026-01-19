package main

func main() {
	LoadConfig()
	EnsureCert()

	SetupRoutes()

	StartServer()
}
