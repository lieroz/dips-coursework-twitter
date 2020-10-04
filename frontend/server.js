const express = require('express');
const bodyParser = require('body-parser');
const axios = require('axios');

const app = express();
const port = 3000;

// Set public folder as root
app.use(express.static('public'));

// Parse POST data as URL encoded data
app.use(bodyParser.urlencoded({
  extended: true,
}));

// Parse POST data as JSON
app.use(bodyParser.json());

// Provide access to node_modules folder
app.use('/scripts', express.static(`${__dirname}/node_modules/`));

const api = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 5000,
});

const refreshApi = axios.create({
  baseURL: 'http://localhost:8000',
  timeout: 5000,
});

async function refreshToken(token) {
  const result = await refreshApi.get('/refresh', {headers: {Cookie: `token=${token}`}});
  if (result.status == 201) {
    return headers['set-cookie'][0].split(";")[0].split("=")[1];
  }
  return token;
}

const errorHandler = (err, req, res) => {
  if (err.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    res.status(403).send({ title: 'Server responded with an error', message: err.message });
  } else if (err.request) {
    // The request was made but no response was received
    res.status(503).send({ title: 'Unable to communicate with server', message: err.message });
  } else {
    // Something happened in setting up the request that triggered an Error
    res.status(500).send({ title: 'An unexpected error occurred', message: err.message });
  }
};

app.post('/api/signup', async (req, res) => {
  try {
    const {username, firstname, lastname, description, password} = req.body;
    const result = await api.post('/signup', { username, firstname, lastname, description, password });
    const {status, headers} = result;

    const token = headers['set-cookie'][0].split(";")[0].split("=")[1];
    res.setHeader('token', token);
    res.status(200).send('');
  } catch (error) {
    errorHandler(error, req, res);
  }
});

app.post('/api/signin', async (req, res) => {
  try {
    const {username, firstname, lastname, description, password} = req.body;
    const result = await api.post('/signin', { username, password });
    const {status, headers} = result;

    const token = headers['set-cookie'][0].split(";")[0].split("=")[1];
    res.setHeader('token', token);
    res.status(200).send('');
  } catch (error) {
    errorHandler(error, req, res);
  }
});

app.get('/api/user/timeline', async (req, res) => {
  try {
    const token = await refreshToken(req.headers.token);

    const username = req.headers.username;
    const result = await api.get('/user/timeline', {headers: {Cookie: `token=${token}`}, data: {username: username}});

    res.setHeader('token', token);
    res.status(200).send(result.data);
  } catch (error) {
    errorHandler(error, req, res);
  }
});

// Redirect all traffic to index.html
app.use((req, res) => res.sendFile(`${__dirname}/public/index.html`));

app.listen(port, () => {
  // eslint-disable-next-line no-console
  console.log('listening on %d', port);
});
