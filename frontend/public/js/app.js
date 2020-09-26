window.addEventListener('load', () => {
  const el = $('#app');

  // Compile Handlebar Templates
  const errorTemplate = Handlebars.compile($('#error-template').html());
  const authTemplate = Handlebars.compile($('#auth-template').html());
  const profileTemplate = Handlebars.compile($('#profile-template').html());
  const homeTemplate = Handlebars.compile($('#home-template').html())

  // Instantiate api handler
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

  router.add('/sign-in', () => {
    const html = authTemplate({
      title: 'Log-in to your account',
      action: 'Login',
      question: 'New to us? <a href="/sign-up">Sign Up</a>',
    });
    el.html(html);

    formValidator();
  });

  router.add('/sign-up', () => {
    const html = authTemplate({
      title: 'Create account ',
      action: 'Create',
      question: 'Already with us? <a href="/sign-in">Sign In</a>',
      fields: `
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="firstname" placeholder="Firstname">
              </div>
            </div>
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="lastname" placeholder="Lastname">
              </div>
            </div>
            <div class="field">
              <div class="ui left icon input">
                <i class="circle icon"></i>
                <input type="text" name="description" placeholder="Description">
              </div>
            </div>
      `,
    });
    el.html(html);

    formValidator();
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

  router.add('/', () => {
    const html = homeTemplate({
      items: {
        1: {
          username: 'lieroz',
          date: 'June 2020',
          text: `Ours is a life of constant reruns. 
            We're always circling back to where we'd we started, then starting all over again. 
            Even if we don't run extra laps that day, we surely will come back for more of the same another day soon.`,
        },
        2: {
          username: 'lieroz',
          date: 'June 2020',
          text: `Ours is a life of constant reruns. 
            We're always circling back to where we'd we started, then starting all over again. 
            Even if we don't run extra laps that day, we surely will come back for more of the same another day soon.`,
        },
        3: {
          username: 'lieroz',
          date: 'June 2020',
          text: `Ours is a life of constant reruns. 
            We're always circling back to where we'd we started, then starting all over again. 
            Even if we don't run extra laps that day, we surely will come back for more of the same another day soon.`,
        },
      },
    });
    el.html(html);
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
