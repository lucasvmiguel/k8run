export function rootHandler(req, res) {
  res.send("Hello Wooooooorld 555!");
}

export function userHandler(req, res) {
  res.send(`Hello, ${req.params.name}!`);
}
