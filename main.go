package main

import(
	"fmt"
	"encoding/json"
	"net/http"
	"log"

	"github.com/graphql-go/graphql"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type person struct{
	Name string
	Email string
}

var sess, err = session.NewSession(
		&aws.Config{
    			 Region: aws.String("ap-south-1"),
    			 },
    			 )
var svc = dynamodb.New(sess)

var persontype = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "person",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},

			"email": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var root = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "root",
		Fields: graphql.Fields{
			"person": &graphql.Field{
				Type: persontype,
				Args: graphql.FieldConfigArgument{
					"Name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error){
						name, isOk := p.Args["Name"].(string)
							if isOk{
								result, err := svc.GetItem(&dynamodb.GetItemInput{
									TableName: aws.String("emails"),
									Key: map[string]*dynamodb.AttributeValue{
										"name": {
											S: aws.String(name),
										},
									},
									})
								if err != nil {
								    fmt.Println(err.Error())
								    return nil,err
								}
								Person := person{}
								err = dynamodbattribute.UnmarshalMap(result.Item, &Person)
								if err != nil {
    								panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
								}

								return Person, nil

							}
						return nil, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: root,
	},
)


func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	log.Print("Executing query\n")
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	return result
}

func Getemails(w http.ResponseWriter, r *http.Request) {
	result := executeQuery(r.URL.Query().Get("query"), schema)
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/email", Getemails)
	fmt.Println("Listening at Localhost:8080")
	http.ListenAndServe(":8080", nil)
}
