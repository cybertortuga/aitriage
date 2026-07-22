from fastapi import BackgroundTasks, FastAPI, Request, Response
from fastapi.responses import HTMLResponse

app = FastAPI()


@app.get("/framework")
def framework_objects(
    request: Request,
    response: Response,
    background_tasks: BackgroundTasks,
):
    return HTMLResponse(f"<p>{request.method} {response.status_code} {background_tasks}</p>")
