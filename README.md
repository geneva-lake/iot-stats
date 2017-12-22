iot-stats is a simple backend application for internet of things devices. It exposes api for register device, report about error and receive firmware.
Also it has admin panel functionality. Administrator can retrieve registered devices list with errors after log in.

For set up one should install dependencies, this can be done by glide:<br />
glide install<br />
Then one should start mongodb and set its uri and credentials to config file.
Also for starting server with TLS files containing a certificate and matching private key must be provided
