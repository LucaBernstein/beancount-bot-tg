import 'package:flutter/material.dart';
import 'package:ui/models/config.dart';

import '../../service/api.dart';

class ConfigWidget extends StatefulWidget {
  final ClientAuthentication authentication;

  const ConfigWidget({super.key, required this.authentication});

  @override
  State<ConfigWidget> createState() => _ConfigWidgetState();
}

class _ConfigWidgetState extends State<ConfigWidget> {
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();

  String? userId;
  String? verificationCode;
  Future<Config?>? config;

  var txtVacation = TextEditingController();
  var txtCurrency = TextEditingController();
  var txtTimezoneOffset = TextEditingController();
  bool isApiEnabled = true;
  bool isOmitLeadingCmdSlash = true;

  void _snackbarMessage(String s) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(s)),
    );
  }

  Future<Config?> _loadConfig() async {
    await widget.authentication.loadExistingToken();
    Config? config;
    String? error;
    (config, error) = await widget.authentication.getConfig();
    if (error != null && error.isNotEmpty) {
      _snackbarMessage(error);
      return null;
    }
    return config;
  }

  Future<void> _saveConfig(Config cnf) async {
    String? error;
    (error,) = await widget.authentication.saveConfig(cnf);
    if (error != null && error.isNotEmpty) {
      _snackbarMessage(error);
    } else {
      _snackbarMessage('Successfully saved configuration.');
    }
  }

  @override
  void initState() {
    super.initState();
    config = _loadConfig();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
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
                        const Text(
                            'Currency (defaults to \'EUR\' if left empty):'),
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
                          validator: (String? value) {
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
                                cnf.currency = txtCurrency.text == ''
                                    ? null
                                    : txtCurrency.text;
                                cnf.timezoneOffset =
                                    int.parse(txtTimezoneOffset.text);
                                cnf.vacationTag = txtVacation.text == ''
                                    ? null
                                    : txtVacation.text;
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
    );
  }
}
