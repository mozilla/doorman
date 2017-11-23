const SERVICE_URL = 'http://localhost:8000'

const AUTH0_CLIENT_ID = 'SLocf7Sa1ibd5GNJMMqO539g7cKvWBOI';
const AUTH0_DOMAIN = 'auth.mozilla.auth0.com';
const AUTH0_CALLBACK_URL = window.location.href;
const SCOPES = 'openid profile';


document.addEventListener('DOMContentLoaded', main);

function main() {
  const logoutBtn = document.getElementById('logout');
  logoutBtn.addEventListener('click', logout);

  const webAuth = new auth0.WebAuth({
    domain: AUTH0_DOMAIN,
    clientID: AUTH0_CLIENT_ID,
    redirectUri: AUTH0_CALLBACK_URL,
    responseType: 'token id_token',
    scope: SCOPES
  });

  // Authentication on Login button
  const loginBtn = document.getElementById('login');
  loginBtn.addEventListener('click', () => {
    webAuth.authorize();
  });

  // New record form.
  const newRecordForm = document.getElementById('api-record-form');
  newRecordForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const formData = new FormData(newRecordForm);
    await postNewRecords(formData.get('name'), formData.get('body'))
    // Empty form once submitted.
    newRecordForm.reset()
  });

  handleAuthentication(webAuth)
}

function handleAuthentication(webAuth) {
  webAuth.parseHash((err, authResult) => {
    if (authResult && authResult.accessToken && authResult.idToken) {
      window.location.hash = '';
      setSession(authResult);
    } else if (err) {
      console.error(err);
      alert(
        'Error: ' + err.error + '. Check the console for further details.'
      );
    } else {
      authResult = JSON.parse(sessionStorage.getItem('session'));
    }

    displayButtons()

    if (isAuthenticated()) {
      console.log('AuthResult', authResult);
      const tokenPayloadDiv = document.getElementById('token-payload');
      tokenPayloadDiv.innerText = JSON.stringify(authResult.idTokenPayload, null, 2);

      Promise.all([
        fetchUserInfo(webAuth),
        showAPIHello(),
        showAPIRecords(),
      ]);
    }
  });
}

function displayButtons() {
  if (isAuthenticated()) {
    document.getElementById('login').setAttribute('disabled', 'disabled');
    document.getElementById('logout').removeAttribute('disabled');
    document.getElementById('view').style.display = 'block';
  } else {
    document.getElementById('login').removeAttribute('disabled');
    document.getElementById('logout').setAttribute('disabled', 'disabled');
    document.getElementById('view').style.display = 'none';
  }
}

function setSession(authResult) {
  // Set the time that the access token will expire at
  const expiresAt = JSON.stringify(
    authResult.expiresIn * 1000 + new Date().getTime()
  );
  sessionStorage.setItem('session', JSON.stringify(authResult));
  sessionStorage.setItem('expires_at', expiresAt);
}

function isAuthenticated() {
  // Check whether the current time is past the
  // access token's expiry time
  const expiresAt = JSON.parse(sessionStorage.getItem('expires_at'));
  return new Date().getTime() < expiresAt;
}

function logout() {
  // Remove tokens and expiry time from sessionStorage
  sessionStorage.removeItem('session');
  sessionStorage.removeItem('expires_at');
  displayButtons();
}

async function fetchUserInfo(webAuth) {
  const auth = JSON.parse(sessionStorage.getItem('session'));
  webAuth.client.userInfo(auth.accessToken, (err, profile) => {
    if (err) {
      console.error(err);
      alert(
        'Error: ' + err.error + '. Check the console for further details.'
      );
    }
    document.getElementById('profile-nickname').innerText = profile.nickname;
    document.getElementById('profile-picture').setAttribute('src', profile.picture);
    document.getElementById('profile-details').innerText = JSON.stringify(profile, null, 2);
  });
}

class APIClient {
  constructor() {
    const auth = JSON.parse(sessionStorage.getItem('session'));
    this.headers = {
      'Authorization': `${auth.tokenType} ${auth.idToken}`,
      'Content-Type': 'application/json'
    };
  }

  async hello() {
    const resp = await fetch(`${SERVICE_URL}/`, {headers: this.headers});
    return await resp.json();
  }

  async list() {
    const resp = await fetch(`${SERVICE_URL}/records`, {headers: this.headers});
    return await resp.json();
  }

  async save(name, body) {
    const resp = await fetch(`${SERVICE_URL}/records/${name}`,
                             {method: 'PUT', body, headers: this.headers});
    return await resp.json();
  }
}

async function showAPIHello() {
  const c = new APIClient();
  const data = await c.hello();

  const apiHelloDiv = document.getElementById('api-hello');
  apiHelloDiv.innerText = JSON.stringify(data, null, 2);
}

async function showAPIRecords() {
  const c = new APIClient();
  const data = await c.list();

  const apiRecordsDiv = document.getElementById('api-records');

  if (data.length == 0) {
    apiRecordsDiv.innerText = 'No records';
    return
  }

  apiRecordsDiv.innerHTML = '';
  for (const {name, body} of data) {
    const _name = document.createElement('h2');
    _name.innerText = name;
    const _body = document.createElement('p');
    _body.className = 'pre';
    _body.innerText = JSON.stringify(body, null, 2);
    apiRecordsDiv.appendChild(_name);
    apiRecordsDiv.appendChild(_body);
  }
}


async function postNewRecords(name, body) {
  const c = new APIClient();
  await c.save(name, body);
  // Refresh list.
  await showAPIRecords();
}
