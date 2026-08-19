// PetVaccineTracker.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

class Vaccine
{
    [JsonPropertyName("name")] public string Name { get; set; }
    [JsonPropertyName("date")] public string Date { get; set; }
    [JsonPropertyName("valid_months")] public int ValidMonths { get; set; }

    [JsonIgnore] public DateTime ParsedDate => DateTime.Parse(Date);
    public DateTime ExpiryDate => ParsedDate.AddMonths(ValidMonths);
    public int DaysUntilExpiry => (int)(ExpiryDate - DateTime.Today).TotalDays;
    public bool IsExpired => DaysUntilExpiry < 0;
}

class Pet
{
    [JsonPropertyName("id")] public int Id { get; set; }
    [JsonPropertyName("name")] public string Name { get; set; }
    [JsonPropertyName("species")] public string Species { get; set; }
    [JsonPropertyName("birthdate")] public string Birthdate { get; set; }
    [JsonPropertyName("owner")] public string Owner { get; set; }
    [JsonPropertyName("vaccines")] public List<Vaccine> Vaccines { get; set; } = new List<Vaccine>();
}

class Tracker
{
    private List<Pet> pets = new List<Pet>();
    private readonly string dataFile = "pets.json";
    private readonly JsonSerializerOptions options = new JsonSerializerOptions { WriteIndented = true };

    public Tracker() { Load(); }

    private void Load()
    {
        if (!File.Exists(dataFile)) return;
        string json = File.ReadAllText(dataFile);
        pets = JsonSerializer.Deserialize<List<Pet>>(json) ?? new List<Pet>();
    }

    private void Save()
    {
        string json = JsonSerializer.Serialize(pets, options);
        File.WriteAllText(dataFile, json);
    }

    private int NextId() => pets.Any() ? pets.Max(p => p.Id) + 1 : 1;

    private Pet GetPet(int id) => pets.FirstOrDefault(p => p.Id == id);

    public void AddPet(string name, string species, string birthdate, string owner)
    {
        var p = new Pet { Id = NextId(), Name = name, Species = species, Birthdate = birthdate, Owner = owner };
        pets.Add(p);
        Save();
        Console.WriteLine($"✅ Pet added: {name} (ID: {p.Id})");
    }

    public void AddVaccine(int petId, string vaccineName, string date, int validMonths)
    {
        var p = GetPet(petId);
        if (p == null) { Console.WriteLine($"❌ Pet with ID {petId} not found."); return; }
        p.Vaccines.Add(new Vaccine { Name = vaccineName, Date = date, ValidMonths = validMonths });
        Save();
        Console.WriteLine($"💉 Vaccine '{vaccineName}' added for {p.Name}.");
    }

    public void ListPets()
    {
        if (!pets.Any()) { Console.WriteLine("No pets."); return; }
        Console.WriteLine("\n🐾 Pets:");
        foreach (var p in pets)
            Console.WriteLine($"  [{p.Id}] {p.Name} ({p.Species}) – Owner: {p.Owner}");
    }

    public void ShowPet(int petId)
    {
        var p = GetPet(petId);
        if (p == null) { Console.WriteLine($"❌ Pet with ID {petId} not found."); return; }
        Console.WriteLine($"\n🐶 \u001b[36m{p.Name}\u001b[0m ({p.Species})");
        Console.WriteLine($"  Owner: {p.Owner}");
        Console.WriteLine($"  Birth: {p.Birthdate}");
        if (!p.Vaccines.Any()) { Console.WriteLine("  No vaccines recorded."); return; }
        Console.WriteLine("  Vaccines:");
        for (int i = 0; i < p.Vaccines.Count; i++)
        {
            var v = p.Vaccines[i];
            int days = v.DaysUntilExpiry;
            string status;
            if (days < 0) status = $"\u001b[31mOVERDUE by {-days} days 🚨\u001b[0m";
            else if (days <= 30) status = $"\u001b[33mexpires in {days} days ⚠️\u001b[0m";
            else status = $"\u001b[32mexpires in {days} days\u001b[0m";
            Console.WriteLine($"    [{i+1}] {v.Name} – {v.Date} – {status}");
        }
    }

    public void CheckExpiring(int daysThreshold = 30)
    {
        var expiring = new List<(Pet pet, Vaccine vaccine, int days)>();
        foreach (var p in pets)
            foreach (var v in p.Vaccines)
            {
                int days = v.DaysUntilExpiry;
                if ((days >= 0 && days <= daysThreshold) || days < 0)
                    expiring.Add((p, v, days));
            }
        if (!expiring.Any()) { Console.WriteLine("✅ No vaccines expiring soon."); return; }
        Console.WriteLine($"\n⏰ Vaccines expiring within {daysThreshold} days (or overdue):");
        foreach (var e in expiring)
        {
            string status = e.days < 0 ? $"\u001b[31mOVERDUE by {-e.days} days\u001b[0m" :
                                         $"\u001b[33m{e.days} days left\u001b[0m";
            Console.WriteLine($"  {e.pet.Name} – {e.vaccine.Name} – {status}");
        }
    }

    static void Main(string[] args)
    {
        if (args.Length < 1) { Console.WriteLine("Usage: PetVaccineTracker <command> [options]"); return; }
        var app = new Tracker();
        string cmd = args[0];
        switch (cmd)
        {
            case "add-pet":
                if (args.Length < 5) { Console.WriteLine("add-pet <name> <species> <birthdate> <owner>"); return; }
                app.AddPet(args[1], args[2], args[3], args[4]);
                break;
            case "add-vaccine":
                if (args.Length < 5) { Console.WriteLine("add-vaccine <pet_id> <vaccine_name> <date> <valid_months>"); return; }
                app.AddVaccine(int.Parse(args[1]), args[2], args[3], int.Parse(args[4]));
                break;
            case "list":
                app.ListPets();
                break;
            case "show":
                if (args.Length < 2) { Console.WriteLine("show <pet_id>"); return; }
                app.ShowPet(int.Parse(args[1]));
                break;
            case "check":
                int days = 30;
                for (int i = 1; i < args.Length; i++)
                    if (args[i] == "--days" && i+1 < args.Length) days = int.Parse(args[++i]);
                app.CheckExpiring(days);
                break;
            default:
                Console.WriteLine("Unknown command.");
                break;
        }
    }
}
