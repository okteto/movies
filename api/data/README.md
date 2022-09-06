# Generating user data with Mimesis

The `users.json` data was generated with [Mimesis](https://mimesis.name/en/master/index.html) and a dataset of north american city/state/zipcodes using the following process.

### Setup

https://mimesis.name/en/master/getting_started.html

- Activate Python virtualenv, and install Mimesis:

  ```
  (env) $ pip install mimesis
  ```

- Download and extract dataset: https://github.com/djbelieny/geoinfo-dataset/blob/master/unique_zipcodes_csv.zip

### Create and run python script

- `user_generate.py`:

  ```python
  from mimesis import Person
  from mimesis.locales import Locale
  from mimesis.enums import Gender

  import random
  import csv
  import json

  genders = ['female', 'male', 'nonbinary', 'fluid']
  gender_distribution = [48, 48, 1, 1] # Sample distribution, adjust to represent your population

  people_genders = random.choices(genders, weights=gender_distribution, k=10000)
  person = Person(Locale.EN)

  with open('unique_zipcodes_csv.csv') as zipcodes:
      csv_rdr = csv.reader(zipcodes)
      # Data is from https://github.com/djbelieny/geoinfo-dataset
      # Raw data headers
      #"city","state","stateISO","country","countryISO","zipCode"
      city_state_zipcode = [
                              [
                                  row[0],
                                  row[1],
                                  row[5]
                              ]
                              for row in csv_rdr]

  users = []
  inc = 0
  for g in people_genders:
      if g == 'female':
          first_name = person.first_name(gender=Gender.FEMALE)
          last_name = person.last_name(gender=Gender.FEMALE)
      elif g == 'male':
          first_name = person.first_name(gender=Gender.MALE)
          last_name = person.last_name(gender=Gender.MALE)
      else:
          first_name = person.first_name()
          last_name = person.last_name()
      city, state, zip_code = city_state_zipcode[random.randint(1,len(city_state_zipcode)-1)]

      telephone = person.telephone(mask="###-###-####")
      inc += 1

      users.append({
              "userId": inc,
              "firstName": first_name,
              "lastName": last_name,
              "phone": telephone,
              "city": city,
              "state": state,
              "zip": zip_code,
              "age": random.randint(13, 109),
              "gender": g
              })


  print(json.dumps(users, indent=4))
  ```

- Run Script

  ```
  (env) $ python user_generate.py > users.json

  ```
