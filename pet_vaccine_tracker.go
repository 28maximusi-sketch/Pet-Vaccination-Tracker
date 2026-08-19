// pet_vaccine_tracker.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Vaccine struct {
	Name        string    `json:"name"`
	Date        time.Time `json:"date"`
	ValidMonths int       `json:"valid_months"`
}

func (v Vaccine) ExpiryDate() time.Time {
	return v.Date.AddDate(0, v.ValidMonths, 0)
}

func (v Vaccine) DaysUntilExpiry() int {
	return int(v.ExpiryDate().Sub(time.Now()).Hours() / 24)
}

func (v Vaccine) IsExpired() bool {
	return v.DaysUntilExpiry() < 0
}

type Pet struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Species   string    `json:"species"`
	Birthdate time.Time `json:"birthdate"`
	Owner     string    `json:"owner"`
	Vaccines  []Vaccine `json:"vaccines"`
}

type Tracker struct {
	Pets []Pet `json:"pets"`
	File string
}

func NewTracker(file string) *Tracker {
	t := &Tracker{File: file}
	t.load()
	return t
}

func (t *Tracker) load() {
	data, err := os.ReadFile(t.File)
	if err != nil {
		return
	}
	json.Unmarshal(data, t)
}

func (t *Tracker) save() {
	data, _ := json.MarshalIndent(t, "", "  ")
	os.WriteFile(t.File, data, 0644)
}

func (t *Tracker) nextID() int {
	max := 0
	for _, p := range t.Pets {
		if p.ID > max {
			max = p.ID
		}
	}
	return max + 1
}

func (t *Tracker) getPet(id int) *Pet {
	for i, p := range t.Pets {
		if p.ID == id {
			return &t.Pets[i]
		}
	}
	return nil
}

func (t *Tracker) AddPet(name, species, birthdate, owner string) {
	birth, _ := time.Parse("2006-01-02", birthdate)
	pet := Pet{
		ID:        t.nextID(),
		Name:      name,
		Species:   species,
		Birthdate: birth,
		Owner:     owner,
	}
	t.Pets = append(t.Pets, pet)
	t.save()
	fmt.Printf("✅ Pet added: %s (ID: %d)\n", name, pet.ID)
}

func (t *Tracker) AddVaccine(petID int, vaccineName, date string, validMonths int) {
	pet := t.getPet(petID)
	if pet == nil {
		fmt.Printf("❌ Pet with ID %d not found.\n", petID)
		return
	}
	vDate, _ := time.Parse("2006-01-02", date)
	v := Vaccine{Name: vaccineName, Date: vDate, ValidMonths: validMonths}
	pet.Vaccines = append(pet.Vaccines, v)
	t.save()
	fmt.Printf("💉 Vaccine '%s' added for %s.\n", vaccineName, pet.Name)
}

func (t *Tracker) ListPets() {
	if len(t.Pets) == 0 {
		fmt.Println("No pets.")
		return
	}
	fmt.Println("\n🐾 Pets:")
	for _, p := range t.Pets {
		fmt.Printf("  [%d] %s (%s) – Owner: %s\n", p.ID, p.Name, p.Species, p.Owner)
	}
}

func (t *Tracker) ShowPet(petID int) {
	pet := t.getPet(petID)
	if pet == nil {
		fmt.Printf("❌ Pet with ID %d not found.\n", petID)
		return
	}
	fmt.Printf("\n🐶 %s (%s)\n", pet.Name, pet.Species)
	fmt.Printf("  Owner: %s\n", pet.Owner)
	fmt.Printf("  Birth: %s\n", pet.Birthdate.Format("2006-01-02"))
	if len(pet.Vaccines) == 0 {
		fmt.Println("  No vaccines recorded.")
		return
	}
	fmt.Println("  Vaccines:")
	for i, v := range pet.Vaccines {
		days := v.DaysUntilExpiry()
		var status string
		if days < 0 {
			status = fmt.Sprintf("\033[31mOVERDUE by %d days 🚨\033[0m", -days)
		} else if days <= 30 {
			status = fmt.Sprintf("\033[33mexpires in %d days ⚠️\033[0m", days)
		} else {
			status = fmt.Sprintf("\033[32mexpires in %d days\033[0m", days)
		}
		fmt.Printf("    [%d] %s – %s – %s\n", i+1, v.Name, v.Date.Format("2006-01-02"), status)
	}
}

func (t *Tracker) CheckExpiring(daysThreshold int) {
	var expiring []struct {
		pet  *Pet
		v    Vaccine
		days int
	}
	for i := range t.Pets {
		for _, v := range t.Pets[i].Vaccines {
			days := v.DaysUntilExpiry()
			if days >= 0 && days <= daysThreshold {
				expiring = append(expiring, struct {
					pet  *Pet
					v    Vaccine
					days int
				}{&t.Pets[i], v, days})
			} else if days < 0 {
				expiring = append(expiring, struct {
					pet  *Pet
					v    Vaccine
					days int
				}{&t.Pets[i], v, days})
			}
		}
	}
	if len(expiring) == 0 {
		fmt.Println("✅ No vaccines expiring soon.")
		return
	}
	fmt.Printf("\n⏰ Vaccines expiring within %d days (or overdue):\n", daysThreshold)
	for _, e := range expiring {
		var status string
		if e.days < 0 {
			status = fmt.Sprintf("\033[31mOVERDUE by %d days\033[0m", -e.days)
		} else {
			status = fmt.Sprintf("\033[33m%d days left\033[0m", e.days)
		}
		fmt.Printf("  %s – %s – %s\n", e.pet.Name, e.v.Name, status)
	}
}

func main() {
	addPetCmd := flag.NewFlagSet("add-pet", flag.ExitOnError)
	addPetName := addPetCmd.String("name", "", "")
	addPetSpecies := addPetCmd.String("species", "", "")
	addPetBirth := addPetCmd.String("birthdate", "", "")
	addPetOwner := addPetCmd.String("owner", "", "")

	addVaccCmd := flag.NewFlagSet("add-vaccine", flag.ExitOnError)
	addVaccPetID := addVaccCmd.Int("pet-id", 0, "")
	addVaccName := addVaccCmd.String("vaccine-name", "", "")
	addVaccDate := addVaccCmd.String("date", "", "")
	addVaccMonths := addVaccCmd.Int("valid-months", 0, "")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	showCmd := flag.NewFlagSet("show", flag.ExitOnError)
	showPetID := showCmd.Int("pet-id", 0, "")
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	checkDays := checkCmd.Int("days", 30, "")

	if len(os.Args) < 2 {
		fmt.Println("Usage: pet_vaccine_tracker <command> [options]")
		return
	}
	tracker := NewTracker("pets.json")

	switch os.Args[1] {
	case "add-pet":
		addPetCmd.Parse(os.Args[2:])
		tracker.AddPet(*addPetName, *addPetSpecies, *addPetBirth, *addPetOwner)
	case "add-vaccine":
		addVaccCmd.Parse(os.Args[2:])
		tracker.AddVaccine(*addVaccPetID, *addVaccName, *addVaccDate, *addVaccMonths)
	case "list":
		listCmd.Parse(os.Args[2:])
		tracker.ListPets()
	case "show":
		showCmd.Parse(os.Args[2:])
		tracker.ShowPet(*showPetID)
	case "check":
		checkCmd.Parse(os.Args[2:])
		tracker.CheckExpiring(*checkDays)
	default:
		fmt.Println("Unknown command")
	}
}
