// pet_vaccine_tracker.cpp
#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <map>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <nlohmann/json.hpp>

using namespace std;
using json = nlohmann::json;

time_t parseDate(const string& s) {
    struct tm tm = {};
    strptime(s.c_str(), "%Y-%m-%d", &tm);
    return mktime(&tm);
}

string formatDate(time_t t) {
    char buf[11];
    strftime(buf, sizeof(buf), "%Y-%m-%d", localtime(&t));
    return string(buf);
}

int daysBetween(time_t a, time_t b) {
    return (int)difftime(a, b) / (24*3600);
}

struct Vaccine {
    string name;
    time_t date;
    int validMonths;

    time_t expiryDate() const {
        struct tm tm = *localtime(&date);
        tm.tm_mon += validMonths;
        return mktime(&tm);
    }

    int daysUntilExpiry() const {
        return daysBetween(expiryDate(), time(nullptr));
    }

    bool isExpired() const { return daysUntilExpiry() < 0; }
};

struct Pet {
    int id;
    string name, species, birthdate, owner;
    vector<Vaccine> vaccines;
};

class Tracker {
private:
    vector<Pet> pets;
    string dataFile = "pets.json";

    void load() {
        ifstream f(dataFile);
        if (!f.is_open()) return;
        json j;
        f >> j;
        for (auto& item : j) {
            Pet p;
            p.id = item["id"];
            p.name = item["name"];
            p.species = item["species"];
            p.birthdate = item["birthdate"];
            p.owner = item["owner"];
            for (auto& v : item["vaccines"]) {
                Vaccine vac;
                vac.name = v["name"];
                vac.date = parseDate(v["date"].get<string>());
                vac.validMonths = v["valid_months"];
                p.vaccines.push_back(vac);
            }
            pets.push_back(p);
        }
    }

    void save() {
        json j = json::array();
        for (auto& p : pets) {
            json vaccines = json::array();
            for (auto& v : p.vaccines) {
                vaccines.push_back({
                    {"name", v.name},
                    {"date", formatDate(v.date)},
                    {"valid_months", v.validMonths}
                });
            }
            j.push_back({
                {"id", p.id},
                {"name", p.name},
                {"species", p.species},
                {"birthdate", p.birthdate},
                {"owner", p.owner},
                {"vaccines", vaccines}
            });
        }
        ofstream f(dataFile);
        f << setw(2) << j << endl;
    }

    int nextId() {
        int max = 0;
        for (auto& p : pets) if (p.id > max) max = p.id;
        return max + 1;
    }

    Pet* getPet(int id) {
        for (auto& p : pets) if (p.id == id) return &p;
        return nullptr;
    }

public:
    Tracker() { load(); }

    void addPet(const string& name, const string& species, const string& birthdate, const string& owner) {
        Pet p{nextId(), name, species, birthdate, owner, {}};
        pets.push_back(p);
        save();
        cout << "✅ Pet added: " << name << " (ID: " << p.id << ")" << endl;
    }

    void addVaccine(int petId, const string& vaccineName, const string& date, int validMonths) {
        Pet* p = getPet(petId);
        if (!p) {
            cout << "❌ Pet with ID " << petId << " not found." << endl;
            return;
        }
        Vaccine v{vaccineName, parseDate(date), validMonths};
        p->vaccines.push_back(v);
        save();
        cout << "💉 Vaccine '" << vaccineName << "' added for " << p->name << "." << endl;
    }

    void listPets() {
        if (pets.empty()) { cout << "No pets." << endl; return; }
        cout << "\n🐾 Pets:" << endl;
        for (auto& p : pets) {
            cout << "  [" << p.id << "] " << p.name << " (" << p.species << ") – Owner: " << p.owner << endl;
        }
    }

    void showPet(int petId) {
        Pet* p = getPet(petId);
        if (!p) {
            cout << "❌ Pet with ID " << petId << " not found." << endl;
            return;
        }
        cout << "\n🐶 \033[36m" << p->name << "\033[0m (" << p->species << ")" << endl;
        cout << "  Owner: " << p->owner << endl;
        cout << "  Birth: " << p->birthdate << endl;
        if (p->vaccines.empty()) {
            cout << "  No vaccines recorded." << endl;
            return;
        }
        cout << "  Vaccines:" << endl;
        for (size_t i = 0; i < p->vaccines.size(); i++) {
            auto& v = p->vaccines[i];
            int days = v.daysUntilExpiry();
            string status;
            if (days < 0) {
                status = "\033[31mOVERDUE by " + to_string(-days) + " days 🚨\033[0m";
            } else if (days <= 30) {
                status = "\033[33mexpires in " + to_string(days) + " days ⚠️\033[0m";
            } else {
                status = "\033[32mexpires in " + to_string(days) + " days\033[0m";
            }
            cout << "    [" << i+1 << "] " << v.name << " – " << formatDate(v.date) << " – " << status << endl;
        }
    }

    void checkExpiring(int daysThreshold = 30) {
        struct Exp { Pet* pet; Vaccine* vac; int days; };
        vector<Exp> expiring;
        for (auto& p : pets) {
            for (auto& v : p.vaccines) {
                int days = v.daysUntilExpiry();
                if ((days >= 0 && days <= daysThreshold) || days < 0) {
                    expiring.push_back({&p, &v, days});
                }
            }
        }
        if (expiring.empty()) {
            cout << "✅ No vaccines expiring soon." << endl;
            return;
        }
        cout << "\n⏰ Vaccines expiring within " << daysThreshold << " days (or overdue):" << endl;
        for (auto& e : expiring) {
            string status = e.days < 0 ? "\033[31mOVERDUE by " + to_string(-e.days) + " days\033[0m" :
                                         "\033[33m" + to_string(e.days) + " days left\033[0m";
            cout << "  " << e.pet->name << " – " << e.vac->name << " – " << status << endl;
        }
    }
};

int main(int argc, char* argv[]) {
    if (argc < 2) {
        cerr << "Usage: pet_vaccine_tracker <command> [options]" << endl;
        return 1;
    }
    Tracker app;
    string cmd = argv[1];
    if (cmd == "add-pet") {
        if (argc < 6) { cerr << "add-pet <name> <species> <birthdate> <owner>" << endl; return 1; }
        app.addPet(argv[2], argv[3], argv[4], argv[5]);
    } else if (cmd == "add-vaccine") {
        if (argc < 6) { cerr << "add-vaccine <pet_id> <vaccine_name> <date> <valid_months>" << endl; return 1; }
        app.addVaccine(stoi(argv[2]), argv[3], argv[4], stoi(argv[5]));
    } else if (cmd == "list") {
        app.listPets();
    } else if (cmd == "show") {
        if (argc < 3) { cerr << "show <pet_id>" << endl; return 1; }
        app.showPet(stoi(argv[2]));
    } else if (cmd == "check") {
        int days = 30;
        for (int i = 2; i < argc; i++) {
            if (string(argv[i]) == "--days" && i+1 < argc) {
                days = stoi(argv[++i]);
            }
        }
        app.checkExpiring(days);
    } else {
        cerr << "Unknown command." << endl;
    }
    return 0;
}
