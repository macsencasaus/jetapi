# jetapi
[JetAPI](https://www.jetapi.dev)

An API to gather information from [JetPhotos](https://www.JetPhotos.com) and [FlightRadar24](https://www.FlightRadar24.com).

## Documentation
See the [documentation](https://www.jetapi.dev/documentation) for more information regarding the usage of the API.

## Getting Started On Your Own
After cloning the project, one can simply run
```
make run
```
to build and run the program.

If you do not have `make`,
```
go run ./cmd/jetapi
```
suffices.

Then one can visit [localhost:8080](http://localhost:8080) to view the documentation and build a query for your local instance.

## More
The API works best with commercial airliners. 
GA aircraft may cause JSON encoding errors due to the variability in FlightRadar's page. 
Registrations not found also return JSON encoding errors.

You may also specify the port and host by setting the `PORT` and `HOST` environment variables respectively:
```
HOST=0.0.0.0 PORT=4000 make run
```
will serve to `0.0.0.0:4000`.
