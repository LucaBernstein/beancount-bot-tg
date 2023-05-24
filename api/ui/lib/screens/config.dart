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
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();

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

  Future<void> _saveConfig(Config cnf) async {
    String? error;
    (error,) = await authentication.saveConfig(cnf);
    if (error != null && error.isNotEmpty) {
      _raiseError(error);
    }
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
        title: Text('Beancount-Bot-Tg Config' /*widget.title*/),
        actions: [
          ElevatedButton(
              onPressed: () => _logout(), child: const Text('Logout')),
        ],
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: SizedBox(
          width: 350,
          child: Form(
              key: _formKey,
              child: FutureBuilder<Config?>(
                future: config,
                builder: (context, snapshot) {
                  if (snapshot.hasData) {
                    txtVacation.text = snapshot.data!.vacationTag ?? '';
                    txtCurrency.text = snapshot.data!.currency ?? '';
                    txtTimezoneOffset.text = '${snapshot.data!.timezoneOffset}';
                    isApiEnabled = snapshot.data!.enableApi;
                    isOmitLeadingCmdSlash =
                        snapshot.data!.omitLeadingCommandSlash;

                    return Column(
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
                        children: [
                          const Text('Vacation tag:'),
                          TextFormField(
                            controller: txtVacation,
                            decoration: const InputDecoration(
                              border: OutlineInputBorder(),
                            ),
                          ),
                          const Text('Currency (defaults to \'EUR\' if left empty):'),
                          TextFormField(
                            controller: txtCurrency,
                            decoration: const InputDecoration(
                              border: OutlineInputBorder(),
                            ),
                          ),
                          const Text('Timezone offset:'),
                          TextFormField(
                            keyboardType: TextInputType.number,
                            controller: txtTimezoneOffset,
                            decoration: const InputDecoration(
                              border: OutlineInputBorder(),
                            ),
                            validator:  (String? value) {
                              String pattern = r'^[0-9]+$';
                              RegExp regExp = RegExp(pattern);
                              String trimmed = (value ?? '').replaceAll(' ', '');
                              if (trimmed == '') {
                                trimmed = '0';
                              }
                              if (regExp.hasMatch(trimmed)) {
                                txtTimezoneOffset.text = trimmed;
                                return null;
                              } else {
                                return 'Not a valid number';
                              }
                            },
                          ),
                          const Text('Enable API:'),
                          Switch(
                            // This bool value toggles the switch.
                            value: isApiEnabled,
                            activeColor: Colors.grey,
                            onChanged: (bool value) {},
                          ),
                          const Text('Omit leading command slash:'),
                          Switch(
                            value: isOmitLeadingCmdSlash,
                            activeColor: Theme.of(context).colorScheme.primary,
                            onChanged: (bool value) {
                              setState(() {
                                (() async {
                                  (await config)!.omitLeadingCommandSlash = value;
                                })();
                              });
                            },
                          ),
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 16.0),
                            child: ElevatedButton(
                              onPressed: () async {
                                if (_formKey.currentState!.validate()) {
                                  _formKey.currentState!.save();
                                  // Update changed config values
                                  Config cnf = (await config)!;
                                  cnf.currency = txtCurrency.text == '' ? null : txtCurrency.text;
                                  cnf.timezoneOffset = int.parse(txtTimezoneOffset.text);
                                  cnf.vacationTag = txtVacation.text == '' ? null : txtVacation.text;
                                  await _saveConfig(cnf);
                                }
                              },
                              child: const Text('Save'),
                            ),
                          )
                        ]);
                  }
                  return const Column();
                },
              )),
        ),
      ),
    );
  }
}
