My app server!

Use the database mysql 5.6

Remember to set the environment variable MARTINI_ENV=production in the production server!

export MARTINI_ENV=production

Generating the Certificates for SSL connection.

openssl genrsa -out key.pem 4096

openssl req -new -nodes -keyout key.key -out server.csr -newkey rsa:2048

Then send the server.csr to our SSLs.com and get the cert.pem, save it in root folder.

Start the server in background.

nohup ./irado &

Nohup will save all log in nohup.out file.

Killing a process running in background.

ps aux | grep irado

,or:

pidof irado

then:

kill PID


