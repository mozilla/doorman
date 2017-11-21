const SERVICE_URL = 'http://iam.local:8000'

const AUTH0_CLIENT_ID = 'SLocf7Sa1ibd5GNJMMqO539g7cKvWBOI';
const AUTH0_DOMAIN = 'auth.mozilla.auth0.com';
const AUTH0_CALLBACK_URL = window.location.href;
const API_AUDIENCE = AUTH0_CLIENT_ID;
const SCOPES = 'openid groups email';


document.addEventListener('DOMContentLoaded', main);

function main() {
  const logoutBtn = document.getElementById('logout');
  logoutBtn.addEventListener('click', logout);

  const webAuth = new auth0.WebAuth({
    domain: AUTH0_DOMAIN,
    clientID: AUTH0_CLIENT_ID,
    redirectUri: AUTH0_CALLBACK_URL,
    audience: API_AUDIENCE,
    responseType: 'token id_token',
    scope: SCOPES
  });

  const loginBtn = document.getElementById('login');
  loginBtn.addEventListener('click', () => {
    webAuth.authorize();
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
      console.log("AuthResult", authResult);
      const tokenPayloadDiv = document.getElementById('token-payload');
      tokenPayloadDiv.innerText = JSON.stringify(authResult.idTokenPayload, null, 2);

      Promise.all([
        fetchUserInfo(webAuth),
        callAPI()
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

async function callAPI() {
  const auth = JSON.parse(sessionStorage.getItem('session'));
  const headers = {
    "Authorization": `${auth.tokenType} ${auth.idToken}`
  };
  const resp = await fetch(`${SERVICE_URL}/`, {headers});
  const data = await resp.json();

  const apiResultDiv = document.getElementById('api-result');
  apiResultDiv.innerText = JSON.stringify(data, null, 2);
}
