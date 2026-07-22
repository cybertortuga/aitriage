from fastapi import FastAPI
from fastapi.responses import Response

app = FastAPI()


@app.get("/charset")
def charset(name: str):
    return Response(f"<h1>{name}</h1>", media_type="text/html; charset=utf-8")
