# go-rmq

This code is pretty much taken verbatim from half way through a PluralSight lecture named
(Go Build Distributed Applications)[http://www.pluralsight.com/courses/go-build-distributed-applications] By Mike Vansickle

I've added comments, and reworded a few things as I've been re-writing the implementation, but
it works.

## Usage

I'm using a docker container as the RabbitMQ broker, which is started with;

    docker run -d --hostname golangrmq --name rmq  -p 5671:5671 -p 5672:5672 -p 15672:15672 rabbitmq:3-management


Start the consumer first, so it hears the discovery mesages from the sensor with;

    go run src/coordinator/exec/main.go -url "amqp://guest:guest@$(docker-machine ip):5672"

Then (any number of) sensors with something like this, where the name is based on the second it was started, so you can start plenty

    go run src/sensor/sensor.go -url "amqp://guest:guest@$(docker-machine ip):5672" -name bob-$(date +%Y%m%d-%H%M%S)


(And any more consumers if you want to...)
