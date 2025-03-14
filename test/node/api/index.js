import express from "express";

import { rootHandler, userHandler } from "./handler/index.js";

const app = express();
const port = 3000;

// Middleware to parse JSON requests
app.use(express.json());

// Home route
app.get("/", rootHandler);

// Start server
app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});
