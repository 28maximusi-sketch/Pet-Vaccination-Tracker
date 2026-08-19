// pet_vaccine_tracker.js
#!/usr/bin/env node
const fs = require('fs');
const { program } = require('commander');
const chalk = require('chalk');

const DATA_FILE = 'pets.json';

class Vaccine {
    constructor(name, date, validMonths) {
        this.name = name;
        this.date = new Date(date);
        this.validMonths = validMonths;
    }

    expiryDate() {
        const d = new Date(this.date);
        d.setMonth(d.getMonth() + this.validMonths);
        return d;
    }

    daysUntilExpiry() {
        const diff = this.expiryDate() - new Date();
        return Math.ceil(diff / (1000 * 60 * 60 * 24));
    }

    isExpired() { return this.daysUntilExpiry() < 0; }

    toJSON() {
        return {
            name: this.name,
            date: this.date.toISOString().slice(0,10),
            validMonths: this.validMonths
        };
    }

    static fromJSON(d) {
        return new Vaccine(d.name, d.date, d.validMonths);
    }
}

class Pet {
    constructor(name, species, birthdate, owner, id = null) {
        this.id = id;
        this.name = name;
        this.species = species;
        this.birthdate = new Date(birthdate);
        this.owner = owner;
        this.vaccines = [];
    }

    toJSON() {
        return {
            id: this.id,
            name: this.name,
            species: this.species,
            birthdate: this.birthdate.toISOString().slice(0,10),
            owner: this.owner,
            vaccines: this.vaccines.map(v => v.toJSON())
        };
    }

    static fromJSON(d) {
        const p = new Pet(d.name, d.species, d.birthdate, d.owner, d.id);
        p.vaccines = (d.vaccines || []).map(v => Vaccine.fromJSON(v));
        return p;
    }
}

class Tracker {
    constructor() {
        this.pets = [];
        this.load();
    }

    load() {
        if (fs.existsSync(DATA_FILE)) {
            const data = JSON.parse(fs.readFileSync(DATA_FILE));
            this.pets = data.map(d => Pet.fromJSON(d));
        }
    }

    save() {
        fs.writeFileSync(DATA_FILE, JSON.stringify(this.pets.map(p => p.toJSON()), null, 2));
    }

    nextId() {
        return this.pets.reduce((max, p) => Math.max(max, p.id || 0), 0) + 1;
    }

    getPet(id) {
        return this.pets.find(p => p.id === id);
    }

    addPet(name, species, birthdate, owner) {
        const pet = new Pet(name, species, birthdate, owner, this.nextId());
        this.pets.push(pet);
        this.save();
        console.log(`✅ Pet added: ${name} (ID: ${pet.id})`);
    }

    addVaccine(petId, vaccineName, date, validMonths) {
        const pet = this.getPet(petId);
        if (!pet) {
            console.log(`❌ Pet with ID ${petId} not found.`);
            return;
        }
        const v = new Vaccine(vaccineName, date, validMonths);
        pet.vaccines.push(v);
        this.save();
        console.log(`💉 Vaccine '${vaccineName}' added for ${pet.name}.`);
    }

    listPets() {
        if (this.pets.length === 0) {
            console.log('No pets.');
            return;
        }
        console.log('\n🐾 Pets:');
        for (const p of this.pets) {
            console.log(`  [${p.id}] ${p.name} (${p.species}) – Owner: ${p.owner}`);
        }
    }

    showPet(petId) {
        const pet = this.getPet(petId);
        if (!pet) {
            console.log(`❌ Pet with ID ${petId} not found.`);
            return;
        }
        console.log(`\n🐶 ${chalk.cyan(pet.name)} (${pet.species})`);
        console.log(`  Owner: ${pet.owner}`);
        console.log(`  Birth: ${pet.birthdate.toISOString().slice(0,10)}`);
        if (pet.vaccines.length === 0) {
            console.log('  No vaccines recorded.');
            return;
        }
        console.log('  Vaccines:');
        for (let i = 0; i < pet.vaccines.length; i++) {
            const v = pet.vaccines[i];
            const days = v.daysUntilExpiry();
            let status;
            if (days < 0) {
                status = chalk.red(`OVERDUE by ${-days} days 🚨`);
            } else if (days <= 30) {
                status = chalk.yellow(`expires in ${days} days ⚠️`);
            } else {
                status = chalk.green(`expires in ${days} days`);
            }
            console.log(`    [${i+1}] ${v.name} – ${v.date.toISOString().slice(0,10)} – ${status}`);
        }
    }

    checkExpiring(daysThreshold = 30) {
        const expiring = [];
        for (const p of this.pets) {
            for (const v of p.vaccines) {
                const days = v.daysUntilExpiry();
                if (days >= 0 && days <= daysThreshold) {
                    expiring.push({ pet: p, vaccine: v, days });
                } else if (days < 0) {
                    expiring.push({ pet: p, vaccine: v, days });
                }
            }
        }
        if (expiring.length === 0) {
            console.log('✅ No vaccines expiring soon.');
            return;
        }
        console.log(`\n⏰ Vaccines expiring within ${daysThreshold} days (or overdue):`);
        for (const e of expiring) {
            const status = e.days < 0 ? chalk.red(`OVERDUE by ${-e.days} days`) : chalk.yellow(`${e.days} days left`);
            console.log(`  ${e.pet.name} – ${e.vaccine.name} – ${status}`);
        }
    }
}

program
    .command('add-pet <name> <species> <birthdate> <owner>')
    .action((name, species, birthdate, owner) => {
        const tracker = new Tracker();
        tracker.addPet(name, species, birthdate, owner);
    });

program
    .command('add-vaccine <petId> <vaccineName> <date> <validMonths>')
    .action((petId, vaccineName, date, validMonths) => {
        const tracker = new Tracker();
        tracker.addVaccine(parseInt(petId), vaccineName, date, parseInt(validMonths));
    });

program
    .command('list')
    .action(() => {
        const tracker = new Tracker();
        tracker.listPets();
    });

program
    .command('show <petId>')
    .action((petId) => {
        const tracker = new Tracker();
        tracker.showPet(parseInt(petId));
    });

program
    .command('check')
    .option('-d, --days <n>', 'Days threshold', parseInt, 30)
    .action((options) => {
        const tracker = new Tracker();
        tracker.checkExpiring(options.days);
    });

program.parse(process.argv);
