🐾 Pet Vaccination Tracker — Multi‑Language Pet Health Manager
8 languages, one complete pet vaccination tracker – manage your pets, track their vaccinations, get reminders for upcoming shots – right from your terminal.

✨ Features
🐶 Add pets – name, species, birth date, owner

💉 Add vaccinations – vaccine name, administration date, validity period (months)

📋 View all pets with their vaccination status

🔍 Check upcoming expirations – list pets with vaccinations expiring soon (default 30 days)

🚨 Highlight overdue vaccinations in red

💾 Persistent storage – all data saved in a local JSON file

📅 Optional reminder – shows days remaining until expiration

🧰 Supported Languages & Files
Language	File	Dependencies
Python	pet_vaccine_tracker.py	none (uses json, datetime, colorama optional)
Go	pet_vaccine_tracker.go	none (stdlib)
JavaScript (Node)	pet_vaccine_tracker.js	chalk (optional)
Ruby	pet_vaccine_tracker.rb	colorize (optional)
PHP	pet_vaccine_tracker.php	none (extensions)
Java	PetVaccineTracker.java	Java 8+ (uses java.time, java.nio)
C#	PetVaccineTracker.cs	.NET Core 3.1+
C++	pet_vaccine_tracker.cpp	nlohmann/json
🚀 Common Usage
All implementations follow the same CLI pattern:

bash
# Add a pet
<command> add-pet "Buddy" "Dog" "2020-05-10" "John"

# Add a vaccine to a pet (use pet ID from list)
<command> add-vaccine <pet_id> "Rabies" "2025-01-15" 36

# List all pets
<command> list

# Show details of a pet (including vaccines)
<command> show <pet_id>

# Check for upcoming expirations (default 30 days)
<command> check

# Check with custom days threshold
<command> check --days 60
Arguments/Commands:

add-pet <name> <species> <birthdate> <owner>

add-vaccine <pet_id> <vaccine_name> <date> <valid_months>

list – show all pets with IDs

show <pet_id> – detailed view with vaccine list

check [--days N] – show vaccines expiring within N days (default 30)

📸 Example Output
text
🐾 Pet Vaccination Tracker

ID: 1 | Buddy (Dog) | Owner: John | Birth: 2020-05-10
  Vaccines:
    [1] Rabies (2025-01-15) – valid 36 months – expires in 15 days ⚠️
    [2] Distemper (2024-12-01) – valid 12 months – OVERDUE! 🚨
📁 Repository Structure
text
.
├── README.md
├── python/
│   └── pet_vaccine_tracker.py
├── go/
│   └── pet_vaccine_tracker.go
├── javascript/
│   └── pet_vaccine_tracker.js
├── ruby/
│   └── pet_vaccine_tracker.rb
├── php/
│   └── pet_vaccine_tracker.php
├── java/
│   └── PetVaccineTracker.java
├── csharp/
│   └── PetVaccineTracker.cs
└── cpp/
    └── pet_vaccine_tracker.cpp
