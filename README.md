# autodev-target

A simple Go playground that **multica-autodev** uses as the codebase its
agents collaborate on. Agents (CEO / Judge / backend-dev / tester / ...)
clone this repo, write code into it, commit to `autodev/issue-*-cycle-*`
branches, push back, and the autodev daemon promotes or rolls back
those branches based on the Judge's score.

This repo is intentionally minimal — its job is to be a playground, not
a production codebase.

## Layout

- `go.mod` — Go module
- `greet/` — sample package agents will be asked to extend
- `.autodev/` — created by agents per cycle (rubric.json, role-plan.json, score.json)

## Notes API

A Go web server with REST API for managing notes, including basic authentication and data persistence.

### Install Dependencies

```bash
go mod tidy
```

### Run the App

```bash
go run main.go
```

The server will start on port 8080.

### Access the UI

Open your browser and navigate to:

```
http://localhost:8080
```

### Authentication

The API requires authentication. To log in:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' \
  -c cookies.txt
```

This will create a session cookie. Use the cookie for subsequent API requests.

### API Endpoints

All API endpoints require authentication (except login).

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | Login with username/password |
| POST | `/api/auth/logout` | Logout and clear session |
| GET | `/api/notes` | List all notes |
| POST | `/api/notes` | Create a new note |
| GET | `/api/notes/:id` | Get a specific note |
| PUT | `/api/notes/:id` | Update a note |
| DELETE | `/api/notes/:id` | Delete a note |

### Example Usage

**Create a note:**

```bash
curl -X POST http://localhost:8080/api/notes \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"title": "My Note", "content": "Note content here", "template": "SOAP"}'
```

**List all notes:**

```bash
curl http://localhost:8080/api/notes -b cookies.txt
```

**Update a note:**

```bash
curl -X PUT http://localhost:8080/api/notes/{id} \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"title": "Updated Title", "content": "Updated content", "template": "SOAP"}'
```

**Delete a note:**

```bash
curl -X DELETE http://localhost:8080/api/notes/{id} -b cookies.txt
```

### Data Persistence

Notes are stored in `notes.json` file. Data persists across app restarts.