# InstaTrack Dashboard

A lightweight, modern, and professional local web dashboard to analyze and track Instagram followers, unfollowers, mutual followers, and recent follow requests. Built completely in **Golang** (no external library dependencies) and **Vanilla HTML/CSS/JS** with a clean zinc-themed responsive design for optimal readability.

## Features

- **Not Following Back (Unfollowers)**: Users you follow who do not follow you back.
- **You Don't Follow Back (Fans)**: Users who follow you but you do not follow back.
- **Mutual Followers**: Users where both follow each other.
- **Recently Unfollowed**: History of profiles you recently unfollowed.
- **Recent Follow Requests**: History of incoming follow requests received.
- **Live Search**: Instant, fuzzy client-side search across all categories.
- **Quick Actions**: One-click "Open Profile" button for each username.
- **Export Utility**: One-click button to copy filtered lists of usernames to your clipboard.
- **Dynamic Reloading**: Reload local data exports instantly from the dashboard without restarting the Go server.

---

## Prerequisites

- **Go** (version 1.22+ recommended) installed on your system.

---

## Getting Started

### 1. Place Data Files
Ensure you place your Instagram data export files (in JSON format) directly into the root directory of this repository:
- `followers_1.json` (also supports multiple files like `followers_2.json` automatically)
- `following.json`
- `recently_unfollowed_profiles.json` (optional)
- `recent_follow_requests.json` (optional)

*Note: These files are already configured in `.gitignore` to prevent you from accidentally committing your personal Instagram data to Git.*

### 2. Run the Server
Start the Go application server:
```bash
go run main.go
```

### 3. Open the Dashboard
Open your browser and navigate to:
[http://localhost:3030](http://localhost:3030)

---

## Project Structure

```text
├── main.go               # Go server, JSON parser, and calculated metrics
├── templates/
│   └── dashboard.html    # Monochromatic zinc-themed layout and UI interactions
├── go.mod                # Go module definition
├── .gitignore            # Git ignore list for personal Instagram data files
└── README.md             # This documentation
```
