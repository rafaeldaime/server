My app server!

It's just an example of an RESTful server in Go, and we also have an REST client in Angularjs.
To run this project, just build the database, running the scripts in database-diagram (an MySQL Workbench diagram), build the server with a go build command, and run it by ./server , you also need to slove the dependencies getting all needed packages by go get command.

Use the database mysql 5.6

Remember to set the environment variable MARTINI_ENV=production in the production server!

export MARTINI_ENV=production

Generating the Certificates for SSL connection.

openssl genrsa -out key.pem 4096

openssl req -new -nodes -keyout key.key -out server.csr -newkey rsa:2048

Then send the server.csr to our SSLs.com and get the cert.pem, save it in root folder.

Start the server in background.

nohup ./server &

Nohup will save all log in nohup.out file.

Killing a process running in background.

ps aux | grep server

,or:

pidof server

then:

kill PID


