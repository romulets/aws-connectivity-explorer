package controller

type JSONResponse struct {
	Status  int
	Content []byte
}

func jsonRes(status int, content []byte) JSONResponse {
	return JSONResponse{status, content}
}
