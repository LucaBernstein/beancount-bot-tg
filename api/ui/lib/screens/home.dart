import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../models/constants.dart';
import '../service/api.dart';
import 'home/config.dart';
import 'home/suggestions.dart';
import 'home/transactions.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  late Future<(bool, String?)> _hasToken;
  ClientAuthentication authentication = ClientAuthentication();

  @override
  void initState() {
    super.initState();
    _hasToken = authentication.loadExistingToken();
    _loadUserInfo();
  }

  _loadUserInfo() async {
    bool isValid;
    (isValid, _) = await _hasToken;
    if (!isValid) {
      _redirectToLogin();
    }
  }

  void _logout() async {
    await authentication.revokeAccess();
    _redirectToLogin();
  }

  void _redirectToLogin() {
    context.go(Routes.login.route);
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<(bool, String?)>(
      future: _hasToken,
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          bool isValid;
          (isValid, _) = snapshot.data!;
          if (isValid) {
            return DefaultTabController(
                initialIndex: 0,
                length: 3,
                child: Scaffold(
                  appBar: AppBar(
                    backgroundColor:
                        Theme.of(context).colorScheme.inversePrimary,
                    // Here we take the value from the MyHomePage object that was created by
                    // the App.build method, and use it to set our appbar title.
                    title: const Text('Beancount-Bot-Tg' /*widget.title*/),
                    actions: [
                      ElevatedButton(
                          onPressed: () => _logout(),
                          child: const Text('Logout')),
                    ],
                    bottom: const TabBar(tabs: <Widget>[
                      Tab(
                        icon: Icon(Icons.monetization_on_outlined),
                        text: 'Transactions',
                      ),
                      Tab(
                        icon: Icon(Icons.list_outlined),
                        text: 'Suggestions',
                      ),
                      Tab(
                        icon: Icon(Icons.settings_suggest_outlined),
                        text: 'Configuration',
                      ),
                    ]),
                  ),
                  body: TabBarView(
                    children: <Widget>[
                      TransactionsWidget(authentication: authentication),
                      SuggestionsWidget(authentication: authentication),
                      ConfigWidget(authentication: authentication),
                    ],
                  ),
                ));
          }
        }
        return const Scaffold(body: Center(child: CircularProgressIndicator()));
      },
    );
  }
}
