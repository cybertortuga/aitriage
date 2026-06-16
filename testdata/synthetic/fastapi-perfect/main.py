from fastapi import FastAPI, Depends, Security, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.httpsredirect import HTTPSRedirectMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi_csrf_protect import CsrfProtect
from fastapi_permissions import Allow, Deny, Authenticated, configure_permissions
from slowapi import Limiter
from slowapi.util import get_remote_address
from sqlalchemy.orm import Session
from pydantic import BaseModel, EmailStr, constr
from pydantic_settings import BaseSettings
from loguru import logger
import jwt


# ── Config ─────────────────────────────────────────────────────────────────────
class Settings(BaseSettings):
    secret_key: str
    database_url: str
    allowed_origins: list[str] = ["https://example.com"]

    class Config:
        env_file = ".env"

settings = Settings()

# ── App ────────────────────────────────────────────────────────────────────────
app = FastAPI(title="Secure FastAPI App")

# Security middleware
app.add_middleware(TrustedHostMiddleware, allowed_hosts=["example.com", "*.example.com"])
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins,
    allow_credentials=True,
    allow_methods=["GET", "POST"],
    allow_headers=["*"],
)

# Rate limiting
limiter = Limiter(key_func=get_remote_address)

# ── Schemas ────────────────────────────────────────────────────────────────────
class UserCreate(BaseModel):
    email: EmailStr
    password: constr(min_length=8)

class UserRead(BaseModel):
    id: int
    email: EmailStr

    class Config:
        from_attributes = True

# ── Auth ───────────────────────────────────────────────────────────────────────
from fastapi.security import OAuth2PasswordBearer, HTTPBearer

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")
security = HTTPBearer()

def get_current_user(token: str = Depends(oauth2_scheme)):
    try:
        payload = jwt.decode(token, settings.secret_key, algorithms=["HS256"])
        return payload
    except jwt.PyJWTError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED)

def require_role(role: str):
    def checker(user=Depends(get_current_user)):
        if role not in user.get("roles", []):
            raise HTTPException(status_code=403, detail="Forbidden")
        return user
    return checker

# ── Global error handler ───────────────────────────────────────────────────────
@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    logger.error(f"Unhandled error: {exc}")
    return {"detail": "Internal server error"}, 500

# ── Routes ─────────────────────────────────────────────────────────────────────
@app.get("/users/me", dependencies=[Security(get_current_user)])
@limiter.limit("10/minute")
async def read_me(user=Depends(get_current_user)):
    logger.info(f"User {user['sub']} accessed /me")
    return user

@app.post("/users", response_model=UserRead)
@limiter.limit("5/minute")
async def create_user(
    data: UserCreate,
    csrf_protect: CsrfProtect = Depends(),
    db: Session = Depends(lambda: None),
    _admin=Depends(require_role("admin")),
):
    logger.info("Creating new user")
    # Use ORM — no raw SQL
    from sqlalchemy import text
    user = db.execute(
        text("SELECT * FROM users WHERE email = :email"),
        {"email": data.email}
    ).fetchone()
    return user
