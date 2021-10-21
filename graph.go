package main1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	//"github.com/dgraph-io/dgo/v200"
	//	"github.com/dgraph-io/dgo/v200/protos/api"
)

type School struct {
	Name  string   `json:"name,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

type loc struct {
	Type   string    `json:"type,omitempty"`
	Coords []float64 `json:"coordinates,omitempty"`
}

// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
// for bool) would be created for values not specified explicitly.

type Person struct {
	Uid      string     `json:"uid,omitempty"`
	Name     string     `json:"name,omitempty"`
	Age      int        `json:"age,omitempty"`
	Dob      *time.Time `json:"dob,omitempty"`
	Married  bool       `json:"married,omitempty"`
	Raw      []byte     `json:"raw_bytes,omitempty"`
	Friends  []Person   `json:"friend,omitempty"`
	Location loc        `json:"loc,omitempty"`
	School   []School   `json:"school,omitempty"`
	DType    []string   `json:"dgraph.type,omitempty"`
}

type CancelFunc func()

func getDgraphClient() (*dgo.Dgraph, CancelFunc) {
	// This example uses dgo
	conn, err := dgo.Dial("https://green-feather-200009.ap-south-1.aws.cloud.dgraph.io/graphql", "NGQyMmRkYzQ0ZDUyMGUwOGI4Mjg0ZjJjMDQxNTA0MmM=")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	//conn, err := grpc.Dial("https://green-feather-200009.ap-south-1.aws.cloud.dgraph.io/graphql", grpc.WithInsecure())
	//	if err != nil {
	//		log.Fatal("While trying to dial gRPC")
	//	}

	//dc := api.NewDgraphClient(conn)
	//	dg := dgo.NewDgraphClient(dc)
	// ctx := context.Background()

	// Perform login call. If the Dgraph cluster does not have ACL and
	// enterprise features enabled, this call should be skipped.
	// for {
	// Keep retrying until we succeed or receive a non-retriable error.
	// err = dg.Login(ctx, “groot”, “password”)
	// if err == nil || !strings.Contains(err.Error(), “Please retry”) {
	// break
	// }
	// time.Sleep(time.Second)
	// }
	// if err != nil {
	// log.Fatalf(“While trying to login %v”, err.Error())
	// }

	return dg, func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error while closing connection:%v", err)
		}
	}
}

func main() {
	dg, cancel := getDgraphClient()
	defer cancel()

	dob := time.Date(1980, 01, 01, 23, 0, 0, 0, time.UTC)
	// While setting an object if a struct has a Uid then its properties in the graph are updated
	// else a new node is created.
	// In the example below new nodes for Alice, Bob and Charlie and school are created (since they
	// don't have a Uid).
	p := Person{
		Uid:     "_:alice",
		Name:    "Alice",
		Age:     26,
		Married: true,
		DType:   []string{"Person"},
		Location: loc{
			Type:   "Point",
			Coords: []float64{1.1, 2},
		},
		Dob: &dob,
		Raw: []byte("raw_bytes"),
		Friends: []Person{{
			Name:  "Bob",
			Age:   24,
			DType: []string{"Person"},
		}, {
			Name:  "Charlie",
			Age:   29,
			DType: []string{"Person"},
		}},
		School: []School{{
			Name:  "Crown Public School",
			DType: []string{"Institution"},
		}},
	}

	op := &api.Operation{}
	op.Schema = `
	  name: string @index(exact) .
	  age: int .
	  married: bool .
	  loc: geo .
	  dob: datetime .
	  Friend: [uid] .
	  type: string .
	  coords: float .
	
	  type Person {
		  name: string
		  age: int
		  married: bool
		  Friend: [Person]
		  loc: Loc
	  }
	
	  type Institution {
		  name: string
	  }
	
	  type Loc {
		  type: string
		  coords: float
	  }`
	ctx := context.Background()
	if err := dg.Alter(ctx, op); err != nil {
		log.Fatal(err)
	}

	mu := &api.Mutation{
		CommitNow: true,
	}
	pb, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	mu.SetJson = pb
	response, err := dg.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}

	// Assigned uids for nodes which were created would be returned in the response.Uids map.
	variables := map[string]string{"$id1": response.Uids["alice"]}
	q := ` query Me($id1: string){
		me(func: uid($id1)) {
			name
			dob
			age
			loc
			raw_bytes
			married
			dgraph.type
			friend @filter(eq(name, "Bob")){
				name
				age
				dgraph.type
			}
			school {
				name
				dgraph.type
			}
		}
		}`

	resp, err := dg.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Me []Person `json:"me"`
	}

	var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)
}
