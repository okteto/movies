const crypto = require("crypto");

const ADMIN_COOKIE = "movies_admin";
const ADMIN_VALUE = "admin";

const secret = () => process.env.SESSION_SECRET || "okteto-movies-demo";

const sign = (value) =>
  crypto.createHmac("sha256", secret()).update(value).digest("hex");

const parseCookies = (header = "") =>
  header.split(";").reduce((cookies, part) => {
    const index = part.indexOf("=");
    if (index > 0) {
      cookies[part.slice(0, index).trim()] = part.slice(index + 1).trim();
    }
    return cookies;
  }, {});

// Validates the admin cookie issued by the api service.
const isAdmin = (req) => {
  const cookie = parseCookies(req.headers.cookie)[ADMIN_COOKIE];
  if (!cookie) {
    return false;
  }

  const separator = cookie.lastIndexOf(".");
  if (separator < 0) {
    return false;
  }

  const value = Buffer.from(cookie.slice(0, separator), "base64url").toString("utf8");
  const signature = Buffer.from(cookie.slice(separator + 1));
  const expected = Buffer.from(sign(value));

  return (
    value === ADMIN_VALUE &&
    signature.length === expected.length &&
    crypto.timingSafeEqual(signature, expected)
  );
};

const requireAdmin = (req, res, next) => {
  if (!isAdmin(req)) {
    res.status(401).json({ error: "admin credentials required" });
    return;
  }
  next();
};

module.exports = { isAdmin, requireAdmin };
