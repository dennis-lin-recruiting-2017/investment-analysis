APP_NAME=investment-analysis
VERSION=0.0.3
UI_PORT?=3000
API_PORT?=8080

.PHONY: frontend backend build clean

frontend:
	cd frontend && npm run build

backend:
	go build -o $(APP_NAME) .

build: frontend backend

clean:
	rm -rf dist frontend/dist
