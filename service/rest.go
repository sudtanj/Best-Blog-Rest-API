package service

import (
	"encoding/json"
	"fmt"
	"gitlab.com/devskiller-tasks/rest-api-blog-golang/model"
	"gitlab.com/devskiller-tasks/rest-api-blog-golang/repository"
	"net/http"
	"strconv"
	"strings"
)

type RestApiService struct {
	postRepository    *repository.PostRepository
	commentRepository *repository.CommentRepository
}

type AckJsonResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func NewRestApiService() RestApiService {
	return RestApiService{postRepository: repository.NewPostRepository(), commentRepository: repository.NewCommentRepository()}
}

func (svc *RestApiService) ServeContent(port int) error {
	portString := ":" + strconv.Itoa(port)
	svc.initializeHandlers()
	return http.ListenAndServe(portString, nil)
}

func (svc *RestApiService) initializeHandlers() {
	http.HandleFunc("/api/post/post", handleAddPost(svc))
	http.HandleFunc("/api/get/post/", handleGetPostByPostId(svc))
	http.HandleFunc("/api/get/comments", handleGetCommentsByPostId(svc))
	http.HandleFunc("/api/post/comment", handleAddComment(svc))
}

func handleAddPost(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var post model.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		if err := svc.postRepository.Insert(post); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			data, err := json.Marshal(&AckJsonResponse{Message: fmt.Sprintf("post id: %d successfully added", post.Id), Status: http.StatusOK})
			if _, err := w.Write(data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

func handleGetPostByPostId(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// example valid api call: GET /api/get/post/42

		// Every responses should have Content-Type=application/json header set

		// should respond with a json response in a format of `AckJsonResponse` with Status 400 and Message "wrong id path variable: PATH_VARIABLE" when invalid ID given,
		// note that also the HTTP response code should be set to 400!
		//eg: GET /api/get/post/abc --> AckJsonResponse{Message: "wrong id path variable: abc", Status: 400}

		// should respond with a json response in a format of `AckJsonResponse` with Status 404 and Message "Post with id: [POST_ID] does not exist"
		// note that also the HTTP response code should be set to 404!
		// when given postID does not exist
		// eg. GET /api/get/post/35 --> '{"Message": "post with id: 35 does not exist", Status: 404}'

		// should respond with valid post entity when post with given id exists:
		// eg. GET /api/get/post/2 --> {"Id": 2, "Title": "test title", "Content": "this is a post content", "CreationDate": "1970-01-01T03:46:40+01:00"}

		vars := strings.Split(r.URL.Path, "/")
		id := vars[len(vars)-1]

		w.Header().Set("Content-Type", "application/json")

		idNum, err := strconv.ParseUint(string(id), 10, 64)

		if err != nil {
			json.NewEncoder(w).Encode(AckJsonResponse{
				Status:  400,
				Message: "wrong id path variable: " + id,
			})
			return
		}

		result, repoErr := svc.postRepository.GetById(idNum)
		if repoErr != nil {
			json.NewEncoder(w).Encode(AckJsonResponse{
				Status:  404,
				Message: "post with id: " + id + " does not exist",
			})
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

func handleGetCommentsByPostId(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// example valid api call: GET /api/get/comments/4

		// Every responses should have Content-Type=application/json header set

		// should respond with a json response in a format of `AckJsonResponse` with Status 400 when invalid ID given, eg: GET /api/get/comments/abc

		// should respond with a valid json response with a list of comments for given postId. If there are no comments for a given postId, should return an empty list
		// eg. example valid api call: GET /api/get/comments/101 -->
		// '[
		//		{"Id": 1, "PostId": 101, "Comment": "comment1", "Author": "author5", "CreationDate" :"1970-01-01T03:46:40+01:05"},
		//		{"Id": 3, "PostId": 101, "Comment": "comment2", "Author": "author4", "CreationDate" :"1970-01-01T03:46:40+01:10"},
		//		{"Id": 5, "PostId": 101, "Comment": "comment3", "Author": "author13", "CreationDate" :"1970-01-01T03:46:40+01:15"}
		//	]'
		vars := strings.Split(r.URL.Path, "/")
		id := vars[len(vars)-1]

		w.Header().Set("Content-Type", "application/json")

		idNum, err := strconv.ParseUint(string(id), 10, 64)
		if err != nil {
			json.NewEncoder(w).Encode(AckJsonResponse{
				Status:  400,
				Message: "wrong id path variable: " + id,
			})
			return
		}

		json.NewEncoder(w).Encode(svc.commentRepository.GetAllByPostId(idNum))
	}
}

func handleAddComment(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// example valid api call: POST /api/post/comment '{"Id": 1, "PostId": 101, "Comment": "comment1", "Author": "author1", "CreationDate" :"1970-01-01T03:46:40+01:00"}'

		// Every responses should have Content-Type=application/json header set

		// should respond with a json response in a format of `AckJsonResponse` with Status code 400 and Message "could not deserialize comment json payload"
		// when invalid or incomplete data posted. Data is considered
		// incomplete when payload misses any member property of the model.
		// Note that HTTP response code also should be 400
		// eg: POST /api/post/comment '{"weird_payload": "weird value"}' --> '{"Message": "could not deserialize comment json payload", Status: 400}'

		// should respond with a json response in a format of `AckJsonResponse` with Status code 400 and json payload Message
		// "Comment with id: COMMENT_ID already exists in the database"
		// when comment with given id already exists in the database.
		// eg: POST /api/post/comment '{"Id": 30, "PostId": 23123, "Comment": "comment1", "Author": "author1", "CreationDate" :"1970-01-01T03:46:40+01:00"}'
		// --> '{"Message": "Comment with id: 30 already exists in the database", Status: 400}'

		// should respond with a json response in a format of `AckJsonResponse` with status code 200 and message 'comment id: COMMENT_ID successfully added' when data was posted successfully.
		// eg: POST /api/post/comment '{"Id": 123, "PostId": 663, "Comment": "this is a comment", "Author": "blogger", "CreationDate" :"1970-01-01T03:46:40+01:00"}' -->
		//'{"Message": "comment id: 123 successfully added", Status: 200}'
		var comment model.Comment
		err := json.NewDecoder(r.Body).Decode(&comment)
		if err != nil {
			json.NewEncoder(w).Encode(AckJsonResponse{
				Status:  400,
				Message: "could not deserialize comment json payload",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		insertErr := svc.commentRepository.Insert(comment)
		if insertErr != nil {
			json.NewEncoder(w).Encode(AckJsonResponse{
				Status:  400,
				Message: "Comment with id: " + strconv.FormatUint(comment.Id, 10) + " already exists in the database",
			})
			return
		}

		json.NewEncoder(w).Encode(&AckJsonResponse{
			Status:  200,
			Message: "comment id: " + strconv.FormatUint(comment.Id, 10) + " successfully added",
		})
	}
}
