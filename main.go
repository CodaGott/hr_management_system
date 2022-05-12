package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017" + dbName

type Employee struct {
	ID string 		`json:"id,omitempty" bson:"_id,omitempty"`
	name string 	`json:"name"`
	email string 	`json:"email"`
	salary float64 	`json:"salary"`
	age int 		`json:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil{
		log.Fatalln("Error occured")
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db: db,
	}
	return nil
}

func main() {
	if err := Connect(); err != nil{
		log.Fatalln(err)
	}
	app := fiber.New()

	app.Get("/employee", func(c *fiber.Ctx) error {
		query := bson.D{{}}
		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		var employees []Employee = make([]Employee, 0)
		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.JSON(employees)
	})
	app.Post("/employee", func(ctx *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		employee := new(Employee)

		if err := ctx.BodyParser(employee); err != nil {
			return ctx.Status(400).SendString(err.Error())
		}
			employee.ID = ""

			insertionResult, err := collection.InsertOne(ctx.Context(), employee)

			if err != nil {
				return ctx.Status(500).SendString(err.Error())
			}
		filer := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(ctx.Context(), filer)

		createdEmp := &Employee{}
		createdRecord.Decode(createdEmp)

		return ctx.Status(200).JSON(createdEmp)
	})
	app.Put("/employee/:id", func(ctx *fiber.Ctx) error {
		idParam := ctx.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			return ctx.SendStatus(400)
		}

		employee := new(Employee)
		if err := ctx.BodyParser(employee); err != nil {
			return ctx.Status(400).SendString(err.Error())
		}
		query := bson.D{{Key: "_id", Value: employeeID}}
		update := bson.D{{
			Key: "$set",
			Value: bson.D{
				{Key: "name", Value: employee.name},
				{Key: "age", Value: employee.age},
				{Key: "email", Value: employee.email},
				{Key: "salary", Value: employee.salary},
			},
		}}
		err = mg.Db.Collection("employees").FindOneAndUpdate(ctx.Context(), query, update).Err()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ctx.SendStatus(400)
			}
			return ctx.SendStatus(500)
		}

		employee.ID = idParam
		return ctx.Status(200).JSON(employee)
	})
	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		employeeID, err := primitive.ObjectIDFromHex(c.Params("id"), )
		if err != nil {
			return c.SendStatus(400)
		}

		query := bson.D{{Key: "_id", Value: employeeID}}
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), &query)

		if err != nil {
			return c.SendStatus(500)
		}

		if result.DeletedCount < 1 {
			return c.SendStatus(404)
		}

		return c.Status(200).JSON("Record deleted")
	})

	log.Fatal(app.Listen(":3000"))
}

