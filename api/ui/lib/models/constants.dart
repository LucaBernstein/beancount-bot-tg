final KeyToken = 'TOKEN';

enum Routes {
  root(route: '/'),
  login(route: '/login'),
  home(route: '/home'),
  config(route: '/config'),
  transactions(route: '/tx');

  const Routes({
    required this.route
  });

  final String route;
}
