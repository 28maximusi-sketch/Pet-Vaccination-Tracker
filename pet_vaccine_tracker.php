# pet_vaccine_tracker.php
<?php
$dataFile = 'pets.json';

class Vaccine {
    public $name, $date, $valid_months;
    function __construct($name, $date, $valid_months) {
        $this->name = $name;
        $this->date = new DateTime($date);
        $this->valid_months = $valid_months;
    }
    function expiryDate() {
        $d = clone $this->date;
        $d->modify("+{$this->valid_months} months");
        return $d;
    }
    function daysUntilExpiry() {
        $now = new DateTime();
        $diff = $this->expiryDate()->diff($now);
        return $diff->invert ? -$diff->days : $diff->days;
    }
    function isExpired() { return $this->daysUntilExpiry() < 0; }
    function toArray() {
        return ['name' => $this->name, 'date' => $this->date->format('Y-m-d'), 'valid_months' => $this->valid_months];
    }
    static function fromArray($d) {
        return new self($d['name'], $d['date'], $d['valid_months']);
    }
}

class Pet {
    public $id, $name, $species, $birthdate, $owner, $vaccines = [];
    function __construct($name, $species, $birthdate, $owner, $id = null) {
        $this->id = $id;
        $this->name = $name;
        $this->species = $species;
        $this->birthdate = new DateTime($birthdate);
        $this->owner = $owner;
    }
    function toArray() {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'species' => $this->species,
            'birthdate' => $this->birthdate->format('Y-m-d'),
            'owner' => $this->owner,
            'vaccines' => array_map(function($v) { return $v->toArray(); }, $this->vaccines)
        ];
    }
    static function fromArray($d) {
        $p = new self($d['name'], $d['species'], $d['birthdate'], $d['owner'], $d['id']);
        if (isset($d['vaccines'])) {
            $p->vaccines = array_map(function($v) { return Vaccine::fromArray($v); }, $d['vaccines']);
        }
        return $p;
    }
}

class Tracker {
    private $pets = [];
    private $file;

    function __construct($file) {
        $this->file = $file;
        $this->load();
    }

    function load() {
        if (file_exists($this->file)) {
            $data = json_decode(file_get_contents($this->file), true);
            foreach ($data as $d) {
                $this->pets[] = Pet::fromArray($d);
            }
        }
    }

    function save() {
        $data = array_map(function($p) { return $p->toArray(); }, $this->pets);
        file_put_contents($this->file, json_encode($data, JSON_PRETTY_PRINT));
    }

    function nextId() {
        $max = 0;
        foreach ($this->pets as $p) if ($p->id > $max) $max = $p->id;
        return $max + 1;
    }

    function getPet($id) {
        foreach ($this->pets as $p) if ($p->id == $id) return $p;
        return null;
    }

    function addPet($name, $species, $birthdate, $owner) {
        $pet = new Pet($name, $species, $birthdate, $owner, $this->nextId());
        $this->pets[] = $pet;
        $this->save();
        echo "✅ Pet added: $name (ID: {$pet->id})\n";
    }

    function addVaccine($petId, $vaccineName, $date, $validMonths) {
        $pet = $this->getPet($petId);
        if (!$pet) {
            echo "❌ Pet with ID $petId not found.\n";
            return;
        }
        $v = new Vaccine($vaccineName, $date, $validMonths);
        $pet->vaccines[] = $v;
        $this->save();
        echo "💉 Vaccine '$vaccineName' added for {$pet->name}.\n";
    }

    function listPets() {
        if (empty($this->pets)) {
            echo "No pets.\n";
            return;
        }
        echo "\n🐾 Pets:\n";
        foreach ($this->pets as $p) {
            echo "  [{$p->id}] {$p->name} ({$p->species}) – Owner: {$p->owner}\n";
        }
    }

    function showPet($petId) {
        $pet = $this->getPet($petId);
        if (!$pet) {
            echo "❌ Pet with ID $petId not found.\n";
            return;
        }
        echo "\n🐶 \033[36m{$pet->name}\033[0m ({$pet->species})\n";
        echo "  Owner: {$pet->owner}\n";
        echo "  Birth: " . $pet->birthdate->format('Y-m-d') . "\n";
        if (empty($pet->vaccines)) {
            echo "  No vaccines recorded.\n";
            return;
        }
        echo "  Vaccines:\n";
        foreach ($pet->vaccines as $i => $v) {
            $days = $v->daysUntilExpiry();
            if ($days < 0) {
                $status = "\033[31mOVERDUE by " . (-$days) . " days 🚨\033[0m";
            } elseif ($days <= 30) {
                $status = "\033[33mexpires in $days days ⚠️\033[0m";
            } else {
                $status = "\033[32mexpires in $days days\033[0m";
            }
            echo "    [" . ($i+1) . "] {$v->name} – " . $v->date->format('Y-m-d') . " – $status\n";
        }
    }

    function checkExpiring($daysThreshold = 30) {
        $expiring = [];
        foreach ($this->pets as $p) {
            foreach ($p->vaccines as $v) {
                $days = $v->daysUntilExpiry();
                if ($days >= 0 && $days <= $daysThreshold) {
                    $expiring[] = ['pet' => $p, 'vaccine' => $v, 'days' => $days];
                } elseif ($days < 0) {
                    $expiring[] = ['pet' => $p, 'vaccine' => $v, 'days' => $days];
                }
            }
        }
        if (empty($expiring)) {
            echo "✅ No vaccines expiring soon.\n";
            return;
        }
        echo "\n⏰ Vaccines expiring within $daysThreshold days (or overdue):\n";
        foreach ($expiring as $e) {
            $status = $e['days'] < 0 ? "\033[31mOVERDUE by " . (-$e['days']) . " days\033[0m" : "\033[33m{$e['days']} days left\033[0m";
            echo "  {$e['pet']->name} – {$e['vaccine']->name} – $status\n";
        }
    }
}

if ($argc < 2) {
    die("Usage: php pet_vaccine_tracker.php <command> [options]\n");
}
$app = new Tracker($dataFile);
$cmd = $argv[1];

switch ($cmd) {
    case 'add-pet':
        if ($argc < 6) die("add-pet <name> <species> <birthdate> <owner>\n");
        $app->addPet($argv[2], $argv[3], $argv[4], $argv[5]);
        break;
    case 'add-vaccine':
        if ($argc < 6) die("add-vaccine <pet_id> <vaccine_name> <date> <valid_months>\n");
        $app->addVaccine((int)$argv[2], $argv[3], $argv[4], (int)$argv[5]);
        break;
    case 'list':
        $app->listPets();
        break;
    case 'show':
        if ($argc < 3) die("show <pet_id>\n");
        $app->showPet((int)$argv[2]);
        break;
    case 'check':
        $days = 30;
        for ($i = 2; $i < $argc; $i++) {
            if ($argv[$i] == '--days' && isset($argv[$i+1])) {
                $days = (int)$argv[$i+1];
                $i++;
            }
        }
        $app->checkExpiring($days);
        break;
    default:
        echo "Unknown command.\n";
}
?>
