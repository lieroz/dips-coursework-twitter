window.addEventListener('load', () => {
  const el = $('#app');

  // Compile Handlebar Templates
  const errorTemplate = Handlebars.compile($('#error-template').html());
  const authTemplate = Handlebars.compile($('#auth-template').html());
  const profileTemplate = Handlebars.compile($('#profile-template').html());
  const homeTemplate = Handlebars.compile($('#home-template').html())

  // Instantiate api handler
  axios.defaults.withCredentials = true
  const api = axios.create({
    baseURL: 'http://localhost:3000/api',
    timeout: 5000,
  });

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
    const { title, message } = error.response.data;
    const html = errorTemplate({ color: 'red', title, message });
    el.html(html);
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
      token = response.headers.token;

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
      token = response.headers.token;

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

  router.add('/profile', () => {
    const html = profileTemplate();
    el.html(html);
  });

  router.add('/following', () => {
    const html = profileTemplate({
      items: {
        '1 Firstname Lastname': 'username',
        '2 Firstname Lastname': 'username',
        '3 Firstname Lastname': 'username',
        '4 Firstname Lastname': 'username',
        '5 Firstname Lastname': 'username',
      },
    });
    el.html(html);
  });

  router.add('/followers', () => {
    const html = profileTemplate({
      items: {
        '6 Firstname Lastname': 'username',
        '7 Firstname Lastname': 'username',
        '8 Firstname Lastname': 'username',
        '9 Firstname Lastname': 'username',
        '10 Firstname Lastname': 'username',
      },
    });
    el.html(html);
  });

  router.add('/tweets', () => {
    const html = profileTemplate({
      items: {
        '1 Tweet title': 'date',
        '2 Tweet title': 'date',
        '3 Tweet title': 'date',
        '4 Tweet title': 'date',
        '5 Tweet title': 'date',
      },
    });
    el.html(html);
  });

  router.add('/', async () => {
    try {
      const username = localStorage.getItem("username");
      const token = localStorage.getItem("token");

      const response = await api.get('/user/timeline', { headers: {"username": username, "token": token} });

      for (obj of response.data) {
        let date = new Date(parseInt(obj.creationTimestamp, 10) * 1000);
        obj['date'] = date.toDateString();
      }

      localStorage.setItem("token", response.headers.token);
      const html = homeTemplate({ items: response.data });
      el.html(html);
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
