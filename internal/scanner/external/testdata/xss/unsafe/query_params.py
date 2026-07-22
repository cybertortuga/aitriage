from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/search")
async def search(request: Request):
    q = request.query_params.get("q")
    return HTMLResponse(content="<p>{}</p>".format(q))
