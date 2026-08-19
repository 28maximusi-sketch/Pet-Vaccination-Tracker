# pet_vaccine_tracker.py
import sys, os, json, argparse
from datetime import datetime, timedelta
from typing import List, Dict, Optional

try:
    from colorama import init, Fore, Style
    init()
    COLORS = True
except ImportError:
    COLORS = False
    Fore = Style = type('', (), {'RESET_ALL':'', 'GREEN':'', 'RED':'', 'YELLOW':'', 'CYAN':''})()

DATA_FILE = "pets.json"

class Vaccine:
    def __init__(self, name: str, date: str, valid_months: int):
        self.name = name
        self.date = datetime.strptime(date, "%Y-%m-%d")
        self.valid_months = valid_months

    def expiry_date(self):
        return self.date + timedelta(days=30*self.valid_months)

    def days_until_expiry(self):
        return (self.expiry_date() - datetime.now()).days

    def is_expired(self):
        return self.days_until_expiry() < 0

    def to_dict(self):
        return {"name": self.name, "date": self.date.isoformat(), "valid_months": self.valid_months}

    @classmethod
    def from_dict(cls, d):
        return cls(d["name"], d["date"], d["valid_months"])

class Pet:
    def __init__(self, name: str, species: str, birthdate: str, owner: str, pet_id: Optional[int] = None):
        self.id = pet_id
        self.name = name
        self.species = species
        self.birthdate = datetime.strptime(birthdate, "%Y-%m-%d")
        self.owner = owner
        self.vaccines: List[Vaccine] = []

    def to_dict(self):
        return {
            "id": self.id,
            "name": self.name,
            "species": self.species,
            "birthdate": self.birthdate.isoformat(),
            "owner": self.owner,
            "vaccines": [v.to_dict() for v in self.vaccines]
        }

    @classmethod
    def from_dict(cls, d):
        p = cls(d["name"], d["species"], d["birthdate"], d["owner"], d["id"])
        p.vaccines = [Vaccine.from_dict(v) for v in d.get("vaccines", [])]
        return p

class Tracker:
    def __init__(self):
        self.pets: List[Pet] = []
        self.load()

    def load(self):
        if os.path.exists(DATA_FILE):
            with open(DATA_FILE, "r") as f:
                data = json.load(f)
                self.pets = [Pet.from_dict(p) for p in data]

    def save(self):
        with open(DATA_FILE, "w") as f:
            json.dump([p.to_dict() for p in self.pets], f, indent=2)

    def next_id(self):
        return max([p.id for p in self.pets] + [0]) + 1

    def get_pet(self, pet_id: int) -> Optional[Pet]:
        for p in self.pets:
            if p.id == pet_id:
                return p
        return None

    def add_pet(self, name, species, birthdate, owner):
        pet = Pet(name, species, birthdate, owner, self.next_id())
        self.pets.append(pet)
        self.save()
        print(f"✅ Pet added: {name} (ID: {pet.id})")

    def add_vaccine(self, pet_id, vaccine_name, date, valid_months):
        pet = self.get_pet(pet_id)
        if not pet:
            print(f"❌ Pet with ID {pet_id} not found.")
            return
        v = Vaccine(vaccine_name, date, valid_months)
        pet.vaccines.append(v)
        self.save()
        print(f"💉 Vaccine '{vaccine_name}' added for {pet.name}.")

    def list_pets(self):
        if not self.pets:
            print("No pets.")
            return
        print(f"\n🐾 {Fore.CYAN}Pets{Style.RESET_ALL}")
        for p in self.pets:
            print(f"  {Fore.YELLOW}[{p.id}]{Style.RESET_ALL} {p.name} ({p.species}) – Owner: {p.owner}")

    def show_pet(self, pet_id):
        pet = self.get_pet(pet_id)
        if not pet:
            print(f"❌ Pet with ID {pet_id} not found.")
            return
        print(f"\n🐶 {Fore.CYAN}{pet.name}{Style.RESET_ALL} ({pet.species})")
        print(f"  Owner: {pet.owner}")
        print(f"  Birth: {pet.birthdate.strftime('%Y-%m-%d')}")
        if not pet.vaccines:
            print("  No vaccines recorded.")
            return
        print("  Vaccines:")
        for i, v in enumerate(pet.vaccines, 1):
            days = v.days_until_expiry()
            if days < 0:
                status = f"{Fore.RED}OVERDUE by {-days} days 🚨{Style.RESET_ALL}"
            elif days <= 30:
                status = f"{Fore.YELLOW}expires in {days} days ⚠️{Style.RESET_ALL}"
            else:
                status = f"{Fore.GREEN}expires in {days} days{Style.RESET_ALL}"
            print(f"    [{i}] {v.name} – {v.date.strftime('%Y-%m-%d')} – {status}")

    def check_expiring(self, days_threshold=30):
        expiring = []
        for p in self.pets:
            for v in p.vaccines:
                days = v.days_until_expiry()
                if 0 <= days <= days_threshold:
                    expiring.append((p, v, days))
                elif days < 0:
                    expiring.append((p, v, days))  # overdue
        if not expiring:
            print("✅ No vaccines expiring soon.")
            return
        print(f"\n⏰ {Fore.CYAN}Vaccines expiring within {days_threshold} days (or overdue){Style.RESET_ALL}")
        for p, v, days in expiring:
            if days < 0:
                status = f"{Fore.RED}OVERDUE by {-days} days{Style.RESET_ALL}"
            else:
                status = f"{Fore.YELLOW}{days} days left{Style.RESET_ALL}"
            print(f"  {p.name} – {v.name} – {status}")

def main():
    parser = argparse.ArgumentParser(description="Pet Vaccination Tracker")
    subparsers = parser.add_subparsers(dest="cmd", required=True)

    add_pet = subparsers.add_parser("add-pet")
    add_pet.add_argument("name")
    add_pet.add_argument("species")
    add_pet.add_argument("birthdate")
    add_pet.add_argument("owner")

    add_vacc = subparsers.add_parser("add-vaccine")
    add_vacc.add_argument("pet_id", type=int)
    add_vacc.add_argument("vaccine_name")
    add_vacc.add_argument("date")
    add_vacc.add_argument("valid_months", type=int)

    subparsers.add_parser("list")
    show_parser = subparsers.add_parser("show")
    show_parser.add_argument("pet_id", type=int)

    check_parser = subparsers.add_parser("check")
    check_parser.add_argument("--days", type=int, default=30)

    args = parser.parse_args()
    tracker = Tracker()

    if args.cmd == "add-pet":
        tracker.add_pet(args.name, args.species, args.birthdate, args.owner)
    elif args.cmd == "add-vaccine":
        tracker.add_vaccine(args.pet_id, args.vaccine_name, args.date, args.valid_months)
    elif args.cmd == "list":
        tracker.list_pets()
    elif args.cmd == "show":
        tracker.show_pet(args.pet_id)
    elif args.cmd == "check":
        tracker.check_expiring(args.days)

if __name__ == "__main__":
    main()
