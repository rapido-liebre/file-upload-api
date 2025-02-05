const express = require("express");
const path = require("path");
const axios = require("axios");

const app = express();
const port = 3000;
const BACKEND_URL = "http://localhost:4000"; // Backend API

// Serve static files (HTML, CSS, JS)
app.use(express.static(path.join(__dirname, "public")));

// Fetch the list of uploaded files
app.get("/files", async (req, res) => {
    try {
        const response = await axios.get(`${BACKEND_URL}/files`);
        res.json(response.data);
    } catch (error) {
        res.status(500).json({ error: "Failed to retrieve files" });
    }
});

// Start the frontend server
app.listen(port, () => {
    console.log(`Frontend running at http://localhost:${port}`);
});
