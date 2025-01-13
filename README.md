# Forum

## Overview

The project is a web forum that allows users to communicate with each other by creating posts and comments. The forum supports associating categories with posts, liking and disliking posts and comments, and filtering posts. The project uses SQLite for data storage, and Docker for containerization.

## Features

- **User Authentication**: Users can register, log in, and log out. User sessions are managed using cookies.
- **Post and Comment Creation**: Registered users can create posts and comments.
- **Categories**: Posts can be associated with one or more categories.
- **Likes and Dislikes**: Registered users can like or dislike posts and comments.
- **Filtering**: Users can filter posts by categories, created posts, and liked posts.
- **Error Handling**: The application handles various HTTP and technical errors gracefully.
- **Docker**: The application is containerized using Docker.

## Technologies Used

- **Go**: The primary programming language used for the backend.
- **SQLite**: The database used for storing user, post, and comment data.
- **Docker**: Used for containerizing the application.
- **HTML/CSS/JavaScript**: Used for the frontend.

## Project Structure

```
.
├── .yaml
├── database/
│   ├── database_test.go
│   └── database.go
├── Dockerfile
├── go.mod
├── go.sum
├── handlers/
│   ├── auth.go
│   ├── category.go
│   ├── comments.go
│   ├── error.go
│   ├── handlers.go
│   ├── middleware.go
│   ├── post.go
│   ├── reactions.go
│   ├── session.go
│   └── utils.go
├── main.go
├── run.sh
├── static/
│   ├── css/
│   │   └── styles.css
│   ├── js/
│   │   ├── categories.js
│   │   ├── comments.js
│   │   ├── dashboard.js
│   │   ├── emailValidation.js
│   │   ├── errorHandler.js
│   │   ├── passwordToggle.js
│   │   ├── passwordValidation.js
│   │   └── postActions.js
├── templates/
│   ├── create-post.html
│   ├── dashboard.html
│   ├── edit-post.html
│   ├── error/
│   │   ├── 400.html
│   │   ├── 401.html
│   │   ├── 403.html
│   │   ├── 404.html
│   │   ├── 405.html
│   │   └── 500.html
│   ├── header.html
│   ├── login.html
│   ├── signup.html
│   └── view-post.html
├── uploads/
└── utils/
    ├── utils_test.go
    └── utils.go
```

## Setup and Installation

### Prerequisites

- Docker
- Go (version 1.22.2 or later)

### Steps

1. **Clone the repository**:
   ```sh
   git clone https://learn.zone01kisumu.ke/git/seodhiambo/forum
   cd forum
   ```

2. **Build and run the Docker container**:
   ```sh
   ./run.sh
   ```

3. **Access the application**:
   Open your web browser and navigate to 

http://localhost:8080

.

## Usage

### User Registration

- Navigate to the signup page.
- Provide your email, username, and password.
- Click "Create account".

### User Login

- Navigate to the login page.
- Provide your username and password.
- Click "Sign in".

### Creating Posts and Comments

- After logging in, navigate to the dashboard.
- Click "Create New Post" to create a post.
- To comment on a post, navigate to the post's page and fill out the comment form.

### Liking and Disliking

- Click the like or dislike buttons on posts and comments to register your reaction.

### Filtering Posts

- Use the sidebar on the dashboard to filter posts by categories.

## Testing

Unit tests are provided for various functionalities. To run the tests, use the following command:

```sh
go test ./...
```

## License

The project is licensed under the [MIT License](#LICENSE).
