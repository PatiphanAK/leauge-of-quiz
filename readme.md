# League of Quiz (Kahoot Clone)

**League of Quiz** is a multiplayer quiz game inspired by Kahoot. Players can join rooms, answer questions, and compete in real-time. The game is built using **Vue3**, **Nuxt3**, **GoFiber**, and **Socket.io** to provide a seamless and interactive gaming experience.

## Technologies Used

- **Frontend**: 
  - Vue 3
  - Nuxt 3
  - Socket.io (for real-time communication)
  
- **Backend**: 
  - GoFiber (Web framework for Go)
  - Socket.io (for real-time communication with clients)

## Features

- **Real-Time Multiplayer**: Players can join quiz rooms and compete with each other in real time.
- **Host Quiz Rooms**: Admins can create and manage quiz rooms, including adding questions and setting timers.
- **Dynamic Question Display**: Questions are displayed to all players in sync, and players submit answers within a set time.
- **Leaderboard**: After each round, players can see their scores in real-time.

## Installation

### Prerequisites

1. Install **Node.js** (for Nuxt3 frontend)
2. Install **Go** (for GoFiber backend)
3. Install **Bun** (for faster frontend dependency management)

### How to setup
Example .env for backend

DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=tatar025
DB_NAME=mydb
DB_PORT=5432

Example Credential Go have to create Creadential and add endpoint at
https://console.cloud.google.com/apis/credentials?referrer=search&project=x-sorter-388112

GOOGLE_CLIENT_ID="12312313131"
GOOGLE_CLIENT_SECRET="123456789"
GOOGLE_REDIRECT_URL="http://localhost:3000/auth/google/callback"


JWT_SECRET="1234HelloKub"
JWT_REFRESH_SECRET="1234Hello"