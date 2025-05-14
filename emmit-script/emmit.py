import polars as pl
import requests
import time 
df = pl.read_csv("weather.dat", separator="\t", has_header=False, new_columns=["date", "temperature", "humidity"])

for row in df.rows():
    requests.post("http://localhost:8080/submit", json={
        "date": row[0],
        "humidity": row[2],
        "temperature": row[1]
        })
    time.sleep(2)
    print(row)
