# 300sf

## Reverse Lookup
https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=59.360814&lon=18.036571&accept-language=sv
https://nominatim.org/release-docs/develop/api/Reverse/



```go
    // Example of detecting if a point is inside a geojson polygon.
	point := geojson.NewPoint(geometry.Point{
		Y: 59.32455958040549,
		X: 18.067933630966028,
	})

	polygon, err := geojson.Parse(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{},"geometry":{"coordinates":[[[18.072986057402943,59.32811225957224],[18.068345175322833,59.32682199593981],[18.065815894589605,59.3268930209002],[18.064493243196836,59.32625379091354],[18.066651253364086,59.324714854592486],[18.065885507820667,59.32361388046536],[18.067416998907532,59.32220505510449],[18.068437992964505,59.322702294252935],[18.0699694840514,59.32237080229547],[18.071245726622863,59.32232344603784],[18.074471139668503,59.32126975223525],[18.075979426344702,59.32285620013397],[18.076791580708544,59.3253185994561],[18.07528329403229,59.32699955806231],[18.072986057402943,59.32811225957224]]],"type":"Polygon"}}]}`, nil)
	if err != nil {
		return
	}

	if polygon.Contains(point) {
		log.Print("Inside")
	} else {
		log.Print("outside")
	}
	return
```
