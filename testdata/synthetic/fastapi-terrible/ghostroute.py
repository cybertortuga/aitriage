from fastapi import FastAPI
app = FastAPI()

@app.get("/ping")
def debug_ping():
    return "pong"
