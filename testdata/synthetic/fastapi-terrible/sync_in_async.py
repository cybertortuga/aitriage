from fastapi import FastAPI
app = FastAPI()

@app.get("/users")
async def get_users():
    db = get_db()
    # Bad! Blocking sync call in async route
    db.commit()
    return {"status": "ok"}
