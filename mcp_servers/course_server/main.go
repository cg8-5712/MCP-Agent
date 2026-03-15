package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// 示例 MCP Server - 课程服务
// 演示如何实现一个符合 MCP-Agent 规范的工具服务

type CallRequest struct {
	StudentID string `json:"student_id"`
	CourseID  string `json:"course_id"`
}

type CallResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Schedule struct {
	StudentID string   `json:"student_id"`
	Courses   []Course `json:"courses"`
}

type Course struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Teacher  string `json:"teacher"`
	Time     string `json:"time"`
	Location string `json:"location"`
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/call", callHandler)

	port := 8081
	log.Printf("Course Server starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func callHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 模拟业务逻辑
	schedule := getStudentSchedule(req.StudentID)

	respondSuccess(w, schedule)
}

func getStudentSchedule(studentID string) Schedule {
	// 模拟数据
	return Schedule{
		StudentID: studentID,
		Courses: []Course{
			{
				ID:       "CS101",
				Name:     "计算机科学导论",
				Teacher:  "张教授",
				Time:     "周一 09:00-11:00",
				Location: "教学楼A101",
			},
			{
				ID:       "MATH201",
				Name:     "高等数学",
				Teacher:  "李教授",
				Time:     "周三 14:00-16:00",
				Location: "教学楼B203",
			},
		},
	}
}

func respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CallResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(CallResponse{
		Code:    status,
		Message: message,
		Data:    nil,
	})
}
