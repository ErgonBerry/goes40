package main

import (
	"context"
	"log"
	"time"
	"strconv"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

type Guest struct {
    Phone         string `bson:"phone"`
    Name          string `bson:"name"`
    Adults        int    `bson:"adults"`
    Children      int    `bson:"children"`
    Confirmed     bool   `bson:"confirmed"`
    IP            string `bson:"ip"`
    OvernightStay string `bson:"overnightStay"` // Novo campo
}

func getTotals() (totalAdults int, totalChildren int, err error) {
    // Pipeline de agregação
    pipeline := []bson.M{
        {
            "$match": bson.M{
                "$or": []bson.M{
                    {"adults": bson.M{"$gt": 0}},   // Filtra documentos onde adults > 0
                    {"children": bson.M{"$gt": 0}}, // Filtra documentos onde children > 0
                },
            },
        },
        {
            "$group": bson.M{
                "_id": nil, // Agrupa todos os documentos
                "totalAdults":   bson.M{"$sum": "$adults"},   // Soma o campo adults
                "totalChildren": bson.M{"$sum": "$children"}, // Soma o campo children
            },
        },
    }

    // Executar a agregação
    cursor, err := collection.Aggregate(context.TODO(), pipeline)
    if err != nil {
        return 0, 0, err
    }
    defer cursor.Close(context.TODO())

    // Ler o resultado
    var result struct {
        TotalAdults   int `bson:"totalAdults"`
        TotalChildren int `bson:"totalChildren"`
    }

    if cursor.Next(context.TODO()) {
        if err := cursor.Decode(&result); err != nil {
            return 0, 0, err
        }
    } else {
        // Se não houver documentos, retorna 0 para ambos
        return 0, 0, nil
    }

    return result.TotalAdults, result.TotalChildren, nil
}

func main() {

	// Carregar a senha do admin
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("A variável de ambiente ADMIN_PASSWORD não está definida")
	}

	// Carregar a URI do MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("A variável de ambiente MONGO_URI não está definida")
	}

	// Conectar ao MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("festa_db").Collection("guests")

	// Configurar o template engine
	engine := html.New("./templates", ".html")

	// Inicializar o Fiber
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Rotas
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("invite", nil)
	})

	app.Get("/step1", func(c *fiber.Ctx) error {
		return c.Render("step1", nil)
	})

	app.Post("/step2", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")
		ip := c.IP() // Captura o IP do usuário

		// Validar se o telefone está no banco de dados
		var guest Guest
		err := collection.FindOne(context.TODO(), bson.M{"phone": phone}).Decode(&guest)
		if err != nil {
			return c.Status(400).SendString("Telefone não encontrado na lista de convidados")
		}

		// Se já confirmou, verificar o IP
		if guest.Confirmed {
			if guest.IP != ip {
				return c.Status(403).SendString("Confirmação já realizada por outro dispositivo.")
			}
			// Se o IP for o mesmo, permitir a alteração
			return c.Render("step2", fiber.Map{
				"Phone":        phone,
				"Adults":       guest.Adults,
				"Children":     guest.Children,
				"IsUpdate":     true, // Indica que é uma atualização
			})
		}

		return c.Render("step2", fiber.Map{
			"Phone": phone,
		})
	})

	app.Post("/step3", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")
		adults := c.FormValue("adults")
		ip := c.IP() // Captura o IP do usuário
	
		// Verifica se o convidado já confirmou presença
		var guest Guest
		err := collection.FindOne(context.TODO(), bson.M{"phone": phone}).Decode(&guest)
		if err != nil {
			return c.Status(400).SendString("Telefone não encontrado na lista de convidados")
		}
	
		// Verifica se é uma alteração
		isUpdate := guest.Confirmed && guest.IP == ip
	
		return c.Render("step3", fiber.Map{
			"Phone":    phone,
			"Adults":   adults,
			"IsUpdate": isUpdate, // Indica se é uma alteração
		})
	})

	app.Post("/step4", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")
		adults := c.FormValue("adults")
		children := c.FormValue("children")
		ip := c.IP() // Captura o IP do usuário
	
		// Verifica se o convidado já confirmou presença
		var guest Guest
		err := collection.FindOne(context.TODO(), bson.M{"phone": phone}).Decode(&guest)
		if err != nil {
			return c.Status(400).SendString("Telefone não encontrado na lista de convidados")
		}
	
		// Verifica se é uma alteração
		isUpdate := guest.Confirmed && guest.IP == ip
	
		return c.Render("step4", fiber.Map{
			"Phone":    phone,
			"Adults":   adults,
			"Children": children,
			"IsUpdate": isUpdate, // Indica se é uma alteração
		})
	})

	app.Post("/confirm", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")
		adults := c.FormValue("adults")
		children := c.FormValue("children")
		overnightStay := c.FormValue("overnightStay") // Novo campo
		ip := c.IP()
	
		// Converter adultos e crianças para inteiros
		adultsInt, err := strconv.Atoi(adults)
		if err != nil {
			return c.Status(400).SendString("Número de adultos inválido")
		}
	
		childrenInt, err := strconv.Atoi(children)
		if err != nil {
			return c.Status(400).SendString("Número de crianças inválido")
		}
	
		// Verificar se o convidado já confirmou
		var guest Guest
		err = collection.FindOne(context.TODO(), bson.M{"phone": phone}).Decode(&guest)
		if err != nil {
			return c.Status(400).SendString("Telefone não encontrado na lista de convidados")
		}
	
		// Se já confirmou, verificar o IP
		if guest.Confirmed {
			if guest.IP != ip {
				return c.Status(403).SendString("Confirmação já realizada por outro dispositivo.")
			}
			// Atualizar os dados no banco
			_, err = collection.UpdateOne(
				context.TODO(),
				bson.M{"phone": phone},
				bson.M{"$set": bson.M{
					"adults":        adultsInt,
					"children":      childrenInt,
					"confirmed":     true,
					"ip":           ip,
					"overnightStay": overnightStay, // Novo campo
				}},
			)
			if err != nil {
				return c.Status(500).SendString("Erro ao atualizar confirmação")
			}
	
			return c.Render("confirm", fiber.Map{
				"Phone":         phone,
				"Adults":        adultsInt,
				"Children":      childrenInt,
				"OvernightStay": overnightStay, // Novo campo
				"IsUpdate":      true,
			})
		}
	
		// Se for uma nova confirmação, atualizar os dados no banco
		_, err = collection.UpdateOne(
			context.TODO(),
			bson.M{"phone": phone},
			bson.M{"$set": bson.M{
				"adults":        adultsInt,
				"children":      childrenInt,
				"confirmed":     true,
				"ip":           ip,
				"overnightStay": overnightStay, // Novo campo
			}},
		)
		if err != nil {
			return c.Status(500).SendString("Erro ao confirmar presença")
		}
	
		return c.Render("confirm", fiber.Map{
			"Phone":         phone,
			"Adults":        adultsInt,
			"Children":      childrenInt,
			"OvernightStay": overnightStay, // Novo campo
			"IsUpdate":      false,
		})
	})

	// Rota para a página de login do admin
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Render("admin_login", nil)
	})

	// Rota para processar o login do admin
	app.Post("/admin/login", func(c *fiber.Ctx) error {
		password := c.FormValue("password")
		if password != adminPassword  {
			return c.Status(401).SendString("Senha incorreta")
		}
		return c.Redirect("/admin/dashboard")
	})

	// Rota para o dashboard do admin
	app.Get("/admin/dashboard", func(c *fiber.Ctx) error {
		// Buscar todos os convidados
		cursor, err := collection.Find(context.TODO(), bson.M{})
		if err != nil {
			return c.Status(500).SendString("Erro ao buscar convidados")
		}
		defer cursor.Close(context.TODO())

		var guests []Guest
		if err := cursor.All(context.TODO(), &guests); err != nil {
			return c.Status(500).SendString("Erro ao decodificar convidados")
		}

		// Calcular totais
		totalAdults, totalChildren, err := getTotals()
		if err != nil {
			return c.Status(500).SendString("Erro ao calcular totais")
		}

		return c.Render("admin_dashboard", fiber.Map{
			"Guests": guests,
			"TotalAdults":   totalAdults,
			"TotalChildren": totalChildren,
		})

	})

	// Rota para adicionar um convidado
	app.Post("/admin/add", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")
		name := c.FormValue("name")

		// Inserir o convidado no banco de dados
		_, err := collection.InsertOne(context.TODO(), bson.M{
			"phone": phone,
			"name":  name,
			"adults": 0,
			"children": 0,
		})
		if err != nil {
			return c.Status(500).SendString("Erro ao adicionar convidado")
		}

		return c.Redirect("/admin/dashboard")
	})

	// Rota para remover um convidado
	app.Post("/admin/delete", func(c *fiber.Ctx) error {
		phone := c.FormValue("phone")

		// Remover o convidado do banco de dados
		_, err := collection.DeleteOne(context.TODO(), bson.M{"phone": phone})
		if err != nil {
			return c.Status(500).SendString("Erro ao remover convidado")
		}

		return c.Redirect("/admin/dashboard")
	})

	// Rota para a página de agradecimento
	app.Get("/thanks", func(c *fiber.Ctx) error {
		return c.Render("thanks", nil)
	})

	// Servir arquivos estáticos
	app.Static("/static", "./static")

	// Iniciar o servidor
	log.Fatal(app.Listen(":3000"))
}