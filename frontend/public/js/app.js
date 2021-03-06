window.addEventListener('load', () => {
  const el = $('#app');

  // Compile Handlebar Templates
  const errorTemplate = Handlebars.compile($('#error-template').html());
  const authTemplate = Handlebars.compile($('#auth-template').html());
  const profileTemplate = Handlebars.compile($('#profile-template').html());
  const profileTemplateTweets = Handlebars.compile($('#profile-template-tweets').html());
  const homeTemplate = Handlebars.compile($('#home-template').html())

  // Instantiate api handler
  axios.defaults.withCredentials = true
  const api = axios.create({
    baseURL: 'http://localhost:3000/api',
    timeout: 30000,
  });

  async function refreshToken() {
    const token = localStorage.getItem('token');
    if (token !== 'undefined') {
      try {
        const result = await api.get('/refresh', {headers: {token: token}});

        if (token != result.headers.token) {
          localStorage.setItem('token', result.headers.token);
        }
      } catch (error) {
        console.log(error);
      }
    }
  }

  setInterval(refreshToken, 5*1000);

  const router = new Router({
    mode: 'history',
    page404: (path) => {
      const html = errorTemplate({
        color: 'yellow',
        title: 'Error 404 - Page NOT Found!',
        message: `The path '/${path}' does not exist on this site`,
      });
      el.html(html);
    },
  });

  // Display Error Banner
  const showError = (error) => {
    console.log(error.response);
    if (error.response.status === 401) {
      router.navigateTo('/signin');
    } else {
      const { title, message } = error.response.data;
      const html = errorTemplate({ color: 'red', title, message });
      el.html(html);
    }
  };

  const formValidator = () => {
    $('.ui.form')
      .form({
        fields: {
          username: {
            identifier  : 'username',
            rules: [
              {
                type   : 'empty',
                prompt : 'Please enter valid username'
              },
              {
                type   : 'length[5]',
                prompt : 'Your username must be at least 5 characters'
              }
            ]
          },
          password: {
            identifier  : 'password',
            rules: [
              {
                type   : 'empty',
                prompt : 'Please enter your password'
              },
              {
                type   : 'length[6]',
                prompt : 'Your password must be at least 6 characters'
              }
            ]
          }
        }
      });
  }

  const signIn = async () => {
    const username = $('#username').val();
    const password = $('#password').val();

    try {
      const response = await api.post('/signin', { username, password });
      const token = response.headers.token;

      localStorage.setItem("username", username);
      localStorage.setItem("token", token);

      router.navigateTo('/');
    } catch (error) {
      showError(error);
    } finally {
      $('.segment').removeClass('loading');
    }
  };

  const signInHandler = () => {
    if ($('.ui.form').form('is valid')) {
      // hide error message
      $('.ui.error.message').hide();
      // Indicate loading status
      $('.segment').addClass('loading');
      signIn();
      // Prevent page from submitting to server
      return false;
    }
    return true;
  };


  router.add('/signin', () => {
    const html = authTemplate({
      title: 'Log-in to your account',
      action: 'Login',
      question: 'New to us? <a href="/signup">Sign Up</a>',
    });
    el.html(html);

    try {
      formValidator();
      $('.submit').click(signInHandler);
    } catch (error) {
        showError(error);
    }
  });

  const signUp = async () => {
    const username = $('#username').val();
    const firstname = $('#firstname').val();
    const lastname = $('#lastname').val();
    const description = $('#description').val();
    const password = $('#password').val();

    try {
      const response = await api.post('/signup', { username, firstname, lastname, description, password });
      const token = response.headers.token;

      localStorage.setItem("username", username);
      localStorage.setItem("token", token);

      router.navigateTo('/');
    } catch (error) {
      showError(error);
    } finally {
      $('.segment').removeClass('loading');
    }
  };

  const signUpHandler = () => {
    if ($('.ui.form').form('is valid')) {
      // hide error message
      $('.ui.error.message').hide();
      // Indicate loading status
      $('.segment').addClass('loading');
      signUp();
      // Prevent page from submitting to server
      return false;
    }
    return true;
  };

  router.add('/signup', async () => {
    const html = authTemplate({
      title: 'Create account ',
      action: 'Create',
      question: 'Already with us? <a href="/signin">Sign In</a>',
      fields: `
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="firstname" id="firstname" placeholder="Firstname">
              </div>
            </div>
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="lastname" id="lastname" placeholder="Lastname">
              </div>
            </div>
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="description" id="description" placeholder="Description">
              </div>
            </div>
      `,
    });
    el.html(html);

    try {
      formValidator();
      $('.submit').click(signUpHandler);
    } catch (error) {
        showError(error);
    }
  });

  router.add('/profile', async () => {
    try {
      const username = localStorage.getItem("username");
      const token = localStorage.getItem("token");

      const response = await api.get('/user/summary', { headers: {"username": username, "token": token} });

      localStorage.setItem("username", response.data.username);
      localStorage.setItem("firstname", response.data.firstname);
      localStorage.setItem("lastname", response.data.lastname);
      localStorage.setItem("description", response.data.description);
      localStorage.setItem("date", response.data.registrationTimestamp);
      localStorage.setItem("tweets", response.data.tweets);

      const date = new Date(parseInt(response.data.registrationTimestamp, 10) * 1000);
      const html = profileTemplate({
        username: response.data.username,
        firstname: response.data.firstname,
        lastname: response.data.lastname,
        description: response.data.description,
        date: date.toDateString(),
      });
      el.html(html);
    } catch (error) {
        showError(error);
    }
  });

  const unfollowUserHandler = async (item) => {
    const followed = item.target.id.split('_').pop();
    const username = localStorage.getItem("username");
    const token = localStorage.getItem("token");

    try {
      const response = await api.post('/user/unfollow', {follower: username, followed: followed}, {headers: {"token": token}});
      router.navigateTo('/following');

    } catch (error) {
      showError(error);
    }
    return false;
  };

  router.add('/following', async () => {
    try {
      const username = localStorage.getItem("username");
      const firstname = localStorage.getItem("firstname");
      const lastname = localStorage.getItem("lastname");
      const description = localStorage.getItem("description");
      const timestamp = localStorage.getItem("date");

      const token = localStorage.getItem("token");
      const response = await api.get('/user/following', { headers: {"username": username, "token": token} });

      const date = new Date(parseInt(timestamp, 10) * 1000);
      const html = profileTemplate({
        username: username,
        firstname: firstname,
        lastname: lastname,
        description: description,
        date: date.toDateString(),
        following: true,
        items: response.data,
      });
      el.html(html);
      $("button[id^='unfollow_user_button_']").click(unfollowUserHandler);
    } catch (error) {
        showError(error);
    }
  });

  router.add('/followers', async () => {
    try {
      const username = localStorage.getItem("username");
      const firstname = localStorage.getItem("firstname");
      const lastname = localStorage.getItem("lastname");
      const description = localStorage.getItem("description");
      const timestamp = localStorage.getItem("date");

      const token = localStorage.getItem("token");
      const response = await api.get('/user/followers', { headers: {"username": username, "token": token} });

      const date = new Date(parseInt(timestamp, 10) * 1000);
      const html = profileTemplate({
        username: username,
        firstname: firstname,
        lastname: lastname,
        description: description,
        date: date.toDateString(),
        items: response.data,
      });
      el.html(html);
    } catch (error) {
        showError(error);
    }
  });

  const deleteTweetsHandler = async (item) => {
    const id = item.target.id.split('_').pop();
    const token = localStorage.getItem("token");

    try {
      const response = await api.post('/tweets/delete', {tweets: [id]}, {headers: {"token": token}});
      router.navigateTo('/tweets');

    } catch (error) {
      showError(error);
    }
    return false;
  };

  router.add('/tweets', async () => {
    try {
      const username = localStorage.getItem("username");
      const firstname = localStorage.getItem("firstname");
      const lastname = localStorage.getItem("lastname");
      const description = localStorage.getItem("description");
      const timestamp = localStorage.getItem("date");

      const date = new Date(parseInt(timestamp, 10) * 1000);

      const tweets = localStorage.getItem("tweets");
      const token = localStorage.getItem("token");
      const response = await api.get('/tweets', {headers: {"tweets": tweets, "token": token} });

      const html = profileTemplateTweets({
        username: username,
        firstname: firstname,
        lastname: lastname,
        description: description,
        date: date.toDateString(),
        items: response.data,
      });

      el.html(html);
      $("button[id^='delete_tweet_button_']").click(deleteTweetsHandler);
    } catch (error) {
        showError(error);
    }
  });

  const postTweetHandler = async () => {
    const content = $('#tweet_textarea').val();
    if (content !== "") {
      $('#tweet_textarea').val('');

      const username = localStorage.getItem("username");
      const token = localStorage.getItem("token");

      try {
        const response = await api.post('/tweets/tweet', {parentId: 0, creator: username, content: content }, {headers: {"token": token}});
      } catch (error) {
        showError(error);
      }
    }
    return false;
  };

  const followUserHandler = async () => {
    const followed = $('#follow_user_input').val();
    if (followed !== "") {
      $('#follow_user_input').val('');

      const username = localStorage.getItem("username");
      const token = localStorage.getItem("token");

      try {
        const response = await api.post('/user/follow', {follower: username, followed: followed}, {headers: {"token": token}});
        router.navigateTo('/');

      } catch (error) {
        showError(error);
      }
    }
    return false;
  };

  const retweetHandler = async (item) => {
    const id = item.target.id.split('_').pop();
    const username = localStorage.getItem("username");
    const token = localStorage.getItem("token");
    const content = document.getElementById(`tweet_content_${id}`).innerText;

    try {
      const response = await api.post('/tweets/tweet', {parentId: id, creator: username, content: content}, {headers: {"token": token}});

    } catch (error) {
      showError(error);
    }
    return false;
  };

  router.add('/', async () => {
    try {
      const username = localStorage.getItem("username");
      const token = localStorage.getItem("token");

      const response = await api.get('/user/timeline', { headers: {"username": username, "token": token} });

      for (obj of response.data) {
        let date = new Date(parseInt(obj.creationTimestamp, 10) * 1000);
        obj['date'] = date.toDateString();
      }

      const html = homeTemplate({ items: response.data });
      el.html(html);

      $('#tweet_button').click(postTweetHandler);
      $('#follow_user_button').click(followUserHandler);
      $("button[id^='retweet_button_']").click(retweetHandler);
    } catch (error) {
      showError(error);
    } finally {
      $('.segment').removeClass('loading');
    }
  });

  router.navigateTo(window.location.pathname);

  // Highlight Active Menu on Load
  const link = $(`a[href$='${window.location.pathname}']`);
  link.addClass('active');

  $('a').on('click', (event) => {
    // Block page load
    event.preventDefault();

    // Highlight Active Menu on Click
    const target = $(event.target);
    $('.item').removeClass('active');
    target.addClass('active');

    // Navigate to clicked url
    const href = target.attr('href');
    const path = href.substr(href.lastIndexOf('/'));
    router.navigateTo(path);
  });
});
