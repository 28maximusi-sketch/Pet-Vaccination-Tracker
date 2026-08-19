// PetVaccineTracker.java
import java.io.*;
import java.nio.file.*;
import java.time.*;
import java.time.format.*;
import java.util.*;
import com.google.gson.*;

class Vaccine {
    String name;
    String date; // ISO
    int validMonths;
    transient LocalDate parsedDate;

    public Vaccine() {}
    public Vaccine(String name, String date, int validMonths) {
        this.name = name;
        this.date = date;
        this.validMonths = validMonths;
        this.parsedDate = LocalDate.parse(date);
    }
    LocalDate expiryDate() { return parsedDate.plusMonths(validMonths); }
    long daysUntilExpiry() { return ChronoUnit.DAYS.between(LocalDate.now(), expiryDate()); }
    boolean isExpired() { return daysUntilExpiry() < 0; }
}

class Pet {
    int id;
    String name, species, birthdate, owner;
    List<Vaccine> vaccines = new ArrayList<>();

    public Pet() {}
    public Pet(int id, String name, String species, String birthdate, String owner) {
        this.id = id;
        this.name = name;
        this.species = species;
        this.birthdate = birthdate;
        this.owner = owner;
    }
}

class Tracker {
    private List<Pet> pets = new ArrayList<>();
    private final String dataFile = "pets.json";
    private final Gson gson = new GsonBuilder().setPrettyPrinting().create();

    public Tracker() { load(); }

    private void load() {
        try {
            Path path = Paths.get(dataFile);
            if (Files.exists(path)) {
                String json = new String(Files.readAllBytes(path));
                Pet[] arr = gson.fromJson(json, Pet[].class);
                pets = Arrays.asList(arr);
            }
        } catch (Exception e) {}
    }

    private void save() {
        try {
            Files.write(Paths.get(dataFile), gson.toJson(pets).getBytes());
        } catch (Exception e) {}
    }

    private int nextId() {
        return pets.stream().mapToInt(p -> p.id).max().orElse(0) + 1;
    }

    private Pet getPet(int id) {
        for (Pet p : pets) if (p.id == id) return p;
        return null;
    }

    public void addPet(String name, String species, String birthdate, String owner) {
        Pet p = new Pet(nextId(), name, species, birthdate, owner);
        pets.add(p);
        save();
        System.out.printf("✅ Pet added: %s (ID: %d)%n", name, p.id);
    }

    public void addVaccine(int petId, String vaccineName, String date, int validMonths) {
        Pet p = getPet(petId);
        if (p == null) {
            System.out.printf("❌ Pet with ID %d not found.%n", petId);
            return;
        }
        Vaccine v = new Vaccine(vaccineName, date, validMonths);
        p.vaccines.add(v);
        save();
        System.out.printf("💉 Vaccine '%s' added for %s.%n", vaccineName, p.name);
    }

    public void listPets() {
        if (pets.isEmpty()) { System.out.println("No pets."); return; }
        System.out.println("\n🐾 Pets:");
        for (Pet p : pets) {
            System.out.printf("  [%d] %s (%s) – Owner: %s%n", p.id, p.name, p.species, p.owner);
        }
    }

    public void showPet(int petId) {
        Pet p = getPet(petId);
        if (p == null) {
            System.out.printf("❌ Pet with ID %d not found.%n", petId);
            return;
        }
        System.out.printf("\n🐶 \033[36m%s\033[0m (%s)%n", p.name, p.species);
        System.out.printf("  Owner: %s%n", p.owner);
        System.out.printf("  Birth: %s%n", p.birthdate);
        if (p.vaccines.isEmpty()) {
            System.out.println("  No vaccines recorded.");
            return;
        }
        System.out.println("  Vaccines:");
        for (int i = 0; i < p.vaccines.size(); i++) {
            Vaccine v = p.vaccines.get(i);
            long days = v.daysUntilExpiry();
            String status;
            if (days < 0) {
                status = String.format("\033[31mOVERDUE by %d days 🚨\033[0m", -days);
            } else if (days <= 30) {
                status = String.format("\033[33mexpires in %d days ⚠️\033[0m", days);
            } else {
                status = String.format("\033[32mexpires in %d days\033[0m", days);
            }
            System.out.printf("    [%d] %s – %s – %s%n", i+1, v.name, v.date, status);
        }
    }

    public void checkExpiring(int daysThreshold) {
        List<Object[]> expiring = new ArrayList<>();
        for (Pet p : pets) {
            for (Vaccine v : p.vaccines) {
                long days = v.daysUntilExpiry();
                if ((days >= 0 && days <= daysThreshold) || days < 0) {
                    expiring.add(new Object[]{p, v, days});
                }
            }
        }
        if (expiring.isEmpty()) {
            System.out.println("✅ No vaccines expiring soon.");
            return;
        }
        System.out.printf("\n⏰ Vaccines expiring within %d days (or overdue):%n", daysThreshold);
        for (Object[] e : expiring) {
            Pet p = (Pet) e[0];
            Vaccine v = (Vaccine) e[1];
            long days = (long) e[2];
            String status = days < 0 ? String.format("\033[31mOVERDUE by %d days\033[0m", -days) :
                                      String.format("\033[33m%d days left\033[0m", days);
            System.out.printf("  %s – %s – %s%n", p.name, v.name, status);
        }
    }

    public static void main(String[] args) throws Exception {
        if (args.length < 1) {
            System.out.println("Usage: PetVaccineTracker <command> [options]");
            return;
        }
        Tracker app = new Tracker();
        String cmd = args[0];
        switch (cmd) {
            case "add-pet":
                if (args.length < 5) { System.out.println("add-pet <name> <species> <birthdate> <owner>"); return; }
                app.addPet(args[1], args[2], args[3], args[4]);
                break;
            case "add-vaccine":
                if (args.length < 5) { System.out.println("add-vaccine <pet_id> <vaccine_name> <date> <valid_months>"); return; }
                app.addVaccine(Integer.parseInt(args[1]), args[2], args[3], Integer.parseInt(args[4]));
                break;
            case "list":
                app.listPets();
                break;
            case "show":
                if (args.length < 2) { System.out.println("show <pet_id>"); return; }
                app.showPet(Integer.parseInt(args[1]));
                break;
            case "check":
                int days = 30;
                for (int i = 1; i < args.length; i++) {
                    if (args[i].equals("--days") && i+1 < args.length) {
                        days = Integer.parseInt(args[i+1]);
                        i++;
                    }
                }
                app.checkExpiring(days);
                break;
            default:
                System.out.println("Unknown command.");
        }
    }
}
