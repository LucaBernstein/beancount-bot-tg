const keyToken = 'TOKEN';

enum Routes {
  root(route: '/'),
  login(route: '/login');

  const Routes({required this.route});

  final String route;
}
