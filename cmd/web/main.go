package main

func main() {
	app := initialize()

	go app.listForShutdown()

	app.serve()
}
