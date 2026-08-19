# pet_vaccine_tracker.rb
require 'json'
require 'date'
require 'colorize'
require 'optparse'

DATA_FILE = 'pets.json'

class Vaccine
  attr_accessor :name, :date, :valid_months

  def initialize(name, date, valid_months)
    @name = name
    @date = Date.parse(date)
    @valid_months = valid_months
  end

  def expiry_date
    @date >> valid_months
  end

  def days_until_expiry
    (expiry_date - Date.today).to_i
  end

  def expired?
    days_until_expiry < 0
  end

  def to_hash
    { name: @name, date: @date.iso8601, valid_months: @valid_months }
  end

  def self.from_hash(h)
    new(h['name'], h['date'], h['valid_months'])
  end
end

class Pet
  attr_accessor :id, :name, :species, :birthdate, :owner, :vaccines

  def initialize(name, species, birthdate, owner, id = nil)
    @id = id
    @name = name
    @species = species
    @birthdate = Date.parse(birthdate)
    @owner = owner
    @vaccines = []
  end

  def to_hash
    {
      id: @id,
      name: @name,
      species: @species,
      birthdate: @birthdate.iso8601,
      owner: @owner,
      vaccines: @vaccines.map(&:to_hash)
    }
  end

  def self.from_hash(h)
    p = new(h['name'], h['species'], h['birthdate'], h['owner'], h['id'])
    p.vaccines = h['vaccines'].map { |v| Vaccine.from_hash(v) }
    p
  end
end

class Tracker
  attr_reader :pets

  def initialize
    @pets = []
    load
  end

  def load
    return unless File.exist?(DATA_FILE)
    data = JSON.parse(File.read(DATA_FILE))
    @pets = data.map { |h| Pet.from_hash(h) }
  end

  def save
    File.write(DATA_FILE, JSON.pretty_generate(@pets.map(&:to_hash)))
  end

  def next_id
    @pets.map(&:id).max.to_i + 1
  end

  def get_pet(id)
    @pets.find { |p| p.id == id }
  end

  def add_pet(name, species, birthdate, owner)
    pet = Pet.new(name, species, birthdate, owner, next_id)
    @pets << pet
    save
    puts "✅ Pet added: #{name} (ID: #{pet.id})"
  end

  def add_vaccine(pet_id, vaccine_name, date, valid_months)
    pet = get_pet(pet_id)
    unless pet
      puts "❌ Pet with ID #{pet_id} not found."
      return
    end
    v = Vaccine.new(vaccine_name, date, valid_months)
    pet.vaccines << v
    save
    puts "💉 Vaccine '#{vaccine_name}' added for #{pet.name}."
  end

  def list_pets
    if @pets.empty?
      puts "No pets."
      return
    end
    puts "\n🐾 Pets:"
    @pets.each do |p|
      puts "  [#{p.id}] #{p.name} (#{p.species}) – Owner: #{p.owner}"
    end
  end

  def show_pet(pet_id)
    pet = get_pet(pet_id)
    unless pet
      puts "❌ Pet with ID #{pet_id} not found."
      return
    end
    puts "\n🐶 #{pet.name.colorize(:cyan)} (#{pet.species})"
    puts "  Owner: #{pet.owner}"
    puts "  Birth: #{pet.birthdate.iso8601}"
    if pet.vaccines.empty?
      puts "  No vaccines recorded."
      return
    end
    puts "  Vaccines:"
    pet.vaccines.each_with_index do |v, i|
      days = v.days_until_expiry
      status = if days < 0
                 "OVERDUE by #{-days} days 🚨".red
               elsif days <= 30
                 "expires in #{days} days ⚠️".yellow
               else
                 "expires in #{days} days".green
               end
      puts "    [#{i+1}] #{v.name} – #{v.date.iso8601} – #{status}"
    end
  end

  def check_expiring(days_threshold = 30)
    expiring = []
    @pets.each do |p|
      p.vaccines.each do |v|
        days = v.days_until_expiry
        if days >= 0 && days <= days_threshold
          expiring << { pet: p, vaccine: v, days: days }
        elsif days < 0
          expiring << { pet: p, vaccine: v, days: days }
        end
      end
    end
    if expiring.empty?
      puts "✅ No vaccines expiring soon."
      return
    end
    puts "\n⏰ Vaccines expiring within #{days_threshold} days (or overdue):"
    expiring.each do |e|
      status = e[:days] < 0 ? "OVERDUE by #{-e[:days]} days".red : "#{e[:days]} days left".yellow
      puts "  #{e[:pet].name} – #{e[:vaccine].name} – #{status}"
    end
  end
end

options = {}
$command = ARGV.shift
if $command.nil?
  puts "Usage: pet_vaccine_tracker.rb <command> [options]"
  exit 1
end

app = Tracker.new

case $command
when "add-pet"
  name, species, birthdate, owner = ARGV.shift(4)
  app.add_pet(name, species, birthdate, owner)
when "add-vaccine"
  pet_id, vaccine_name, date, valid_months = ARGV.shift(4)
  app.add_vaccine(pet_id.to_i, vaccine_name, date, valid_months.to_i)
when "list"
  app.list_pets
when "show"
  pet_id = ARGV.shift.to_i
  app.show_pet(pet_id)
when "check"
  days = 30
  if ARGV.include?('--days')
    idx = ARGV.index('--days')
    days = ARGV[idx+1].to_i if idx
  end
  app.check_expiring(days)
else
  puts "Unknown command."
end
