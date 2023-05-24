import 'package:flutter/material.dart';
import 'package:ui/models/config.dart';

import '../models/constants.dart';
import '../service/api.dart';

class ConfigPage extends StatefulWidget {
  const ConfigPage({super.key});

  @override
  State<ConfigPage> createState() => _ConfigPageState();
}

class _ConfigPageState extends State<ConfigPage> {
  String? userId;
  String? verificationCode;
  ClientAuthentication authentication = ClientAuthentication();
  Future<Config?>? config;

  var txtVacation = TextEditingController();
  var txtCurrency = TextEditingController();
  var txtTimezoneOffset = TextEditingController();
  bool isApiEnabled = true;
  bool isOmitLeadingCmdSlash = true;

  void _raiseError(String error) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(error)),
    );
  }

  Future<Config?> _loadConfig() async {
    await authentication.loadExistingToken();
    Config? config;
    String? error;
    (config, error) = await authentication.getConfig();
    if (error != null && error.isNotEmpty) {
      _raiseError(error);
      return null;
    }
    return config;
  }

  void _logout() async {
    await authentication.revokeAccess();
    _redirectToLogin();
  }

  void _redirectToLogin() {
    Navigator.pushNamedAndRemoveUntil(
        context, Routes.login.route, (Route<dynamic> route) => false);
  }

  @override
  void initState() {
    super.initState();
    config = _loadConfig();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text('Beancount-Bot-Tg Web-UI Home' /*widget.title*/),
        actions: [ElevatedButton(onPressed: () => _logout(), child: const Text('Logout')),],
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Column(
          // Column is also a layout widget. It takes a list of children and
          // arranges them vertically. By default, it sizes itself to fit its
          // children horizontally, and tries to be as tall as its parent.
          //
          // Column has various properties to control how it sizes itself and
          // how it positions its children. Here we use mainAxisAlignment to
          // center the children vertically; the main axis here is the vertical
          // axis because Columns are vertical (the cross axis would be
          // horizontal).
          //
          // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
          // action in the IDE, or press "p" in the console), to see the
          // wireframe for each widget.
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text('Beancount-Bot-Tg configuration values:'),
            FutureBuilder<Config?>(
              future: config,
              builder: (context, snapshot) {
                if (snapshot.hasData) {
                  txtVacation.text = snapshot.data!.vacationTag ?? '';
                  txtCurrency.text = snapshot.data!.currency ?? '';
                  txtTimezoneOffset.text = '${snapshot.data!.timezoneOffset}';
                  isApiEnabled = snapshot.data!.enableApi;
                  isOmitLeadingCmdSlash = snapshot.data!.omitLeadingCommandSlash;

                  return Column(children: [
                    const Text('Vacation tag:'),
                    SizedBox(
                      width: 250,
                      child: TextField(
                        controller: txtVacation,
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                        ),
                        onEditingComplete: () => {print('editing vacation complete: ${txtVacation.text}')},
                      ),
                    ),
                    const Text('Currency:'),
                    SizedBox(
                      width: 250,
                      child: TextField(
                        controller: txtCurrency,
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                        ),
                        onEditingComplete: () => {print('editing currency complete: ${txtCurrency.text}')},
                      ),
                    ),
                    const Text('Timezone offset:'),
                    SizedBox(
                      width: 250,
                      child: TextField(
                        keyboardType: TextInputType.number,
                        controller: txtTimezoneOffset,
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                        ),
                        onEditingComplete: () => {print('editing timezone complete: ${txtTimezoneOffset.text}')},
                      ),
                    ),
                    const Text('Enable API:'),
                    Switch(
                      // This bool value toggles the switch.
                      value: isApiEnabled,
                      activeColor: Colors.red,
                      onChanged: (bool value) {},
                    ),
                    const Text('Omit leading command slash:'),
                    Switch(
                      // This bool value toggles the switch.
                      value: isOmitLeadingCmdSlash,
                      activeColor: Colors.red,
                      onChanged: (bool value) {
                        setState(() {
                          isOmitLeadingCmdSlash = value;
                        });
                        print(isOmitLeadingCmdSlash);
                        snapshot.data!.omitLeadingCommandSlash = !snapshot.data!.omitLeadingCommandSlash;
                      },
                    )
                  ]);
                }
                return const Column();
              },
            )
          ],
        ),
      ),
    );
  }
}
