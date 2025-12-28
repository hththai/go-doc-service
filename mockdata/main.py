import csv
from faker import Faker

fake = Faker()
num_rows = 500_000
file_name = 'fake_data.csv'

with open(file_name, mode='w', newline='') as file:
    writer = csv.writer(file)
    # Write header (optional, but good practice)
    writer.writerow(['user_name', 'password', 'created_at'])

    for _ in range(num_rows):
        writer.writerow([
            fake.unique.user_name(),
            fake.password(length=20, special_chars=True, digits=True, upper_case=True, lower_case=True),
        ])

print(f"Generated {num_rows} rows in {file_name}")
