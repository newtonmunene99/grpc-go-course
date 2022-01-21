package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/newtonmunene99/grpc-go-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var blogCollection *mongo.Collection

type server struct {
	blogpb.BlogServiceServer
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {

	fmt.Printf("Create blog request: %v\n", req)

	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetContent(),
		Content:  blog.GetTitle(),
	}

	res, err := blogCollection.InsertOne(context.Background(), data)

	if err != nil {
		log.Fatalf("Failed to insert to DB: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			"Cannot convert to OID",
		)
	}

	resp := &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}

	return resp, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {

	fmt.Printf("Read blog request: %v\n", req)

	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Cannot parse ID",
		)
	}

	data := &blogItem{}
	filter := bson.D{
		{
			Key:   "_id",
			Value: oid,
		},
	}

	findErr := blogCollection.FindOne(context.Background(), filter).Decode(data)

	if findErr == mongo.ErrNoDocuments {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %s", blogID),
		)
	}

	if findErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", findErr),
		)
	}

	resp := &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}

	return resp, nil

}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Printf("Update blog request: %v\n", req)

	blog := req.GetBlog()
	blogID := blog.GetId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID: %v", blogID),
		)
	}

	data := &blogItem{}
	filter := bson.D{
		{
			Key:   "_id",
			Value: oid,
		},
	}

	findErr := blogCollection.FindOne(context.Background(), filter).Decode(data)

	if findErr == mongo.ErrNoDocuments {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %s", blogID),
		)
	}

	if findErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", findErr),
		)
	}

	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	updateRes, updateErr := blogCollection.ReplaceOne(context.Background(), filter, data)

	if updateErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", updateErr),
		)
	}

	if updateRes.ModifiedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %s", blogID),
		)
	}

	resp := &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}

	return resp, nil

}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {

	fmt.Printf("Delete blog request: %v\n", req)

	blogID := req.GetBlogId()

	oid, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Cannot parse ID",
		)
	}

	filter := bson.D{
		{
			Key:   "_id",
			Value: oid,
		},
	}

	deleteRes, deleteErr := blogCollection.DeleteOne(context.Background(), filter)

	if deleteErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", deleteErr),
		)
	}

	if deleteRes.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %s", blogID),
		)
	}

	resp := &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}

	return resp, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {

	fmt.Printf("List blog request: %v\n", req)

	filter := bson.D{}

	cur, err := blogCollection.Find(context.Background(), filter)

	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}

		decodeErr := cur.Decode(data)

		if decodeErr != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data: %v", decodeErr),
			)
		}

		resp := &blogpb.ListBlogResponse{
			Blog: &blogpb.Blog{
				Id:       data.ID.Hex(),
				AuthorId: data.AuthorID,
				Title:    data.Title,
				Content:  data.Content,
			},
		}

		stream.Send(resp)
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	return nil
}

func main() {
	// If we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog Server Started")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	fmt.Println("Connecting to Mongo")

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatalf("Failed to connect to mongo: %v", err)
	}

	defer func() {
		fmt.Println("Closing Mongo connection")

		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Connected to MongoDB")

	blogCollection = mongoClient.Database("grpc-go-course").Collection("blog")

	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	tls := false
	opts := []grpc.ServerOption{}

	if tls {
		certFile := "ssl/server.crt"

		keyFile := "ssl/server.pem"

		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)

		if sslErr != nil {
			log.Fatalf("Fialed loading certificates: %v\n", sslErr)
			return
		}

		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)

	blogpb.RegisterBlogServiceServer(s, &server{})

	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server")

		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}

	}()

	// Wait for exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch

	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()

	fmt.Println("Server stopped")
}
