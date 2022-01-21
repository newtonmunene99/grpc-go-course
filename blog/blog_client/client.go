package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/newtonmunene99/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	fmt.Println("Blog client")

	tls := false
	opts := grpc.WithInsecure()

	if tls {

		certFile := "ssl/ca.crt"

		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")

		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
			return
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	// blog := createNewBlog(c)

	// readBlog(c, blog.Blog.Id)

	// updateBlog(c, &blogpb.Blog{
	// 	Id:       blog.Blog.Id,
	// 	AuthorId: "Newton",
	// 	Title:    "Third blog",
	// 	Content:  "Content of the third blog",
	// })

	// deleteBlog(c, blog.Blog.Id)

	listBlog(c)
}

func createNewBlog(c blogpb.BlogServiceClient) *blogpb.CreateBlogResponse {

	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Newton",
			Title:    "Second blog",
			Content:  "Content of the second blog",
		},
	}

	res, err := c.CreateBlog(context.Background(), req)

	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Blog has been created: %v\n", res)

	return res
}

func readBlog(c blogpb.BlogServiceClient, id string) {

	req := &blogpb.ReadBlogRequest{
		BlogId: id,
	}

	res, err := c.ReadBlog(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while reading blog: %v\n", err)
		return
	}

	fmt.Printf("Blog was read: %v\n", res)
}

func updateBlog(c blogpb.BlogServiceClient, blog *blogpb.Blog) {

	req := &blogpb.UpdateBlogRequest{
		Blog: blog,
	}

	res, err := c.UpdateBlog(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while updating blog: %v\n", err)
		return
	}

	fmt.Printf("Blog was updated: %v\n", res)
}

func deleteBlog(c blogpb.BlogServiceClient, id string) {

	req := &blogpb.DeleteBlogRequest{
		BlogId: id,
	}

	res, err := c.DeleteBlog(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while deleting blog: %v\n", err)
		return
	}

	fmt.Printf("Blog was deleted: %v\n", res)
}

func listBlog(c blogpb.BlogServiceClient) {

	req := &blogpb.ListBlogRequest{}

	stream, err := c.ListBlog(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while listing blog: %v\n", err)
		return
	}

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while reading stream: %v\n", err)
			return
		}

		blog := res.GetBlog()

		fmt.Printf("Blog was read: %v\n", blog)
	}
}
