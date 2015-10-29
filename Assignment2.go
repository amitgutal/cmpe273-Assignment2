package main

import (
	"gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
    "github.com/julienschmidt/httprouter"
    "io/ioutil"
    "fmt"
	"strings"
	"net/http"
	"encoding/json"
)

type GoogleCoordinates struct {
	
	Results []struct {
		AddressComponents []struct {
			LongName string `json:"long_name"`
			ShortName string `json:"short_name"`
			Types []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
		Types []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type Response struct
{
  Id     bson.ObjectId `json:"id" bson:"_id"`
  Name string	`json:"name" bson:"name"`
  Address string	`json:"address" bson:"address" `
  City string		`json:"city"  bson:"city"`
  State string	`json:"state"  bson:"state"`
  ZipCode string	`json:"zip"  bson:"zip" `

Coordinate struct 
  {
       Lat float64 `json:"lat"   bson:"lat"`
       Lng float64 `json:"lng"   bson:"lng"`
  } `json:"coordinate" bson:"coordinate"`
}


type MongoSessionObject struct {  

    session *mgo.Session
}
   


	  
func LocationDetails(respWriter http.ResponseWriter, httpRequest *http.Request, params httprouter.Params) {  
    
 mongoSession := GetMongoSession(dbConnection())

 object_iden:= params.ByName("object_iden")

if !bson.IsObjectIdHex(object_iden) {
	
	    fmt.Println(" Page Not Found Error .. 404")
        respWriter.WriteHeader(404)
        return
    }
    
   bson_Object := bson.ObjectIdHex(object_iden)
    
   db_resp := Response{}
    
   db_err := mongoSession.session.DB("cmpe273").C("users").FindId(bson_Object).One(&db_resp); 

    if db_err != nil {
	
        respWriter.WriteHeader(404)
		fmt.Println(" Page Not Found Error .. 404")
        return
    }
    
    
    marshal_Obj, _ := json.MarshalIndent(db_resp, "", " ")
    
    respWriter.Header().Set("Content-Type", "application/json")
    respWriter.WriteHeader(200)
    fmt.Fprintf(respWriter, "%s", marshal_Obj)
	    
}

func LocationUpdate (respWriter http.ResponseWriter, request *http.Request, params httprouter.Params) {  
    
var coordinate GoogleCoordinates

var response_put Response

	object_iden:= params.ByName("object_iden")

	if !bson.IsObjectIdHex(object_iden) {
		
        respWriter.WriteHeader(404)
		fmt.Println("Record Not Found")
        return
		
    }
  
  httpClient := &http.Client{}
  
  bsonObjectId := bson.ObjectIdHex(object_iden)
    
  json.NewDecoder(request.Body).Decode(&response_put)   

  Address_Url :=  response_put.Address+","+response_put.City+","+response_put.State+","+response_put.ZipCode

  Address_Url = strings.Replace(Address_Url," ","+",-1)

  Address_Url = "https://maps.google.com/maps/api/geocode/json?address="+Address_Url+"&sensor=false"
    
new_Req, _ := http.NewRequest("GET", Address_Url, nil)

resp,_:= httpClient.Do(new_Req)

if( resp.StatusCode >= 200 && resp.StatusCode < 300 ) {
	
	fmt.Println("Status Code between 200 and 300")
    body, _ := ioutil.ReadAll(resp.Body) 
    
	_= json.Unmarshal(body, &coordinate)

   }
     
    for _,Results := range coordinate.Results {
		
    	    response_put.Coordinate.Lat= Results.Geometry.Location.Lat
	        response_put.Coordinate.Lng = Results.Geometry.Location.Lng
         }
            
    db_session := GetMongoSession(dbConnection()) 

    db_session.session.DB("cmpe273").C("users").Update(bson.M{ "_id":bsonObjectId},bson.M{"$set":bson.M{"address":response_put.Address,"city":response_put.City,"state":response_put.State,"zip":response_put.ZipCode,"coordinate.lat":response_put.Coordinate.Lat,"coordinate.lng":response_put.Coordinate.Lng}})
   
    connection_err := db_session.session.DB("cmpe273").C("users").FindId(bsonObjectId).One(&response_put); 
	
	   if connection_err != nil {
		
         respWriter.WriteHeader(404)
         fmt.Println("Page Not Found ..Error 404")
		 return
		
      }
    
    
	
    marshal_obj, _ := json.Marshal(response_put)
	
	
    respWriter.Header().Set("Content-Type", "application/json")
    respWriter.WriteHeader(201)
    fmt.Fprintf(respWriter, "%s", marshal_obj)
	
 }

 func LocationCreate (response_Writer http.ResponseWriter, resp_Request *http.Request, _ httprouter.Params) {  
   
   var googleCoordinate GoogleCoordinates
   var response_post Response
    
  json.NewDecoder(resp_Request.Body).Decode(&response_post)

  Google_Api:=  response_post.Address+","+response_post.City+","+response_post.State+","+response_post.ZipCode
  Google_Api = strings.Replace(Google_Api," ","+",-1)
  Google_Api = "https://maps.google.com/maps/api/geocode/json?address="+Google_Api+"&sensor=false"
    
client := &http.Client{}

req, _ := http.NewRequest("GET", Google_Api, nil)

resp,_:= client.Do(req)

if( resp.StatusCode >= 200 && resp.StatusCode < 300 ) {
	
          body, _ := ioutil.ReadAll(resp.Body) 
          _= json.Unmarshal(body, &googleCoordinate)
     }
     
    for _,Sample := range googleCoordinate.Results {
		
    	    response_post.Coordinate.Lat = Sample.Geometry.Location.Lat
	        response_post.Coordinate.Lng = Sample.Geometry.Location.Lng
	
         }
             response_post.Id = bson.NewObjectId()
		
	db_session := GetMongoSession(dbConnection()) 	
			
    db_session.session.DB("cmpe273").C("users").Insert(response_post)
	
	resp_Writer, _ := json.MarshalIndent(response_post, "", "    ")
   
   fmt.Fprintf(response_Writer, "%s", resp_Writer)
   response_Writer.WriteHeader(201)
}
	  
	  
func LocationDelete (respW http.ResponseWriter, httpReq *http.Request, param httprouter.Params) {  


  
  object_iden := param.ByName("object_iden")

    
    if !bson.IsObjectIdHex(object_iden) {
		
		fmt.Println(" Page Not Found Error .. 404")
        respW.WriteHeader(404)
        return
		
    }

    mongoSession := GetMongoSession(dbConnection())
    bsonObject := bson.ObjectIdHex(object_iden)

   
   db_error := mongoSession.session.DB("cmpe273").C("users").RemoveId(bsonObject); 
	
	if db_error != nil {
		
		fmt.Println(" Page Not Found Error")
        respW.WriteHeader(404)
        return
    }

   fmt.Println("Page Found Response 200")
   respW.WriteHeader(200)
}	  

	  	  
			
	func GetMongoSession(Mongo_Session *mgo.Session) *MongoSessionObject {
	
    return &MongoSessionObject{Mongo_Session}
}		

func dbConnection() *mgo.Session { 
 
    // Connect to our local mongo
	
    Mongo_Session, Mongo_err := mgo.Dial("mongodb://amit:1234@ds031407.mongolab.com:31407/cmpe273")

    // Check if connection error, is mongo running?
	
    if Mongo_err != nil {
		
		fmt.Println("Error in connection to Mongo Instance")
        panic(Mongo_err)
    }
	
    return Mongo_Session
}
		  
func main() {


    
    multiplexer := httprouter.New()
  
    multiplexer.PUT("/locations/:object_iden",LocationUpdate)
    
    multiplexer.POST("/locations",LocationCreate)
	
    multiplexer.DELETE("/locations/:object_iden", LocationDelete)
    
	multiplexer.GET("/locations/:object_iden", LocationDetails)


    Serv_Object := http.Server{
       
	   Addr:        "localhost:8080",
       Handler: multiplexer,
    }
    Serv_Object.ListenAndServe()
	
	fmt.Println("Server Running")
}