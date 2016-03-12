# go-rmq

This code is pretty much taken verbatim from half way through a PluralSight lecture named
[Go Build Distributed Applications](http://www.pluralsight.com/courses/go-build-distributed-applications) By Mike Vansickle

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

## Help!

The main scripts have `-help` options you can use;

### Coordinators

      go run coordinator/exec/main.go -help
      Usage of /var/folders/9q/2mbtqfx911q8h8jj2xhk6jxw0000gn/T/go-build162413432/command-line-arguments/_obj/exe/main:
        -rmq_url string
          	RabbitMQ endpoint URL (default "amqp://guest:guest@192.168.99.100:5672")
      exit status 2

### Sensors

      go run sensor/sensor.go  -help
     Usage of /var/folders/9q/2mbtqfx911q8h8jj2xhk6jxw0000gn/T/go-build030101474/command-line-arguments/_obj/exe/sensor:
       -freq uint
         	update frequency in cycles/sec (default 5)
       -max float
         	maximum value for generated readings (default 5)
       -min float
         	minimum value for generated readings (default 1)
       -name string
         	name of the sensor (default "sensor")
       -rmq_url string
         	RabbitMQ endpoint URL (default "amqp://guest:guest@192.168.99.100:5672")
       -step float
         	maximum allowable change per measurement (default 0.1)
     exit status 2
