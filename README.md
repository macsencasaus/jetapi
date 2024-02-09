# jetapi
A RESTful API to gather information from [JetPhotos](https://www.JetPhotos.com) and [FlightRadar24](https://www.FlightRadar24.com).

## Getting Started
Clone 
```
git clone https://github.com/macsencasaus/jetapi.git
```
Then
```
go run .
```
Will automatically serve on port :8080.

## Query
For now, only one query string is available which is the registration of the aircraft under the `/api` fixed path.
For example:
```
http://localhost:8080/api?reg=g-xlea
```
This will return information from both sites including images and prior flights!

## More
Works best with commercial airliners. GA aircraft may cause JSON encoding errors due to the variability in FlightRadar's page. Registrations not found also return JSON encoding errors.

