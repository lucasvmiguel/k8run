export function rootHandler(req, res) {
  res.send("Hello World!");
}

export function userHandler(req, res) {
  res.send(`Hello, ${req.params.name}!`);
}
