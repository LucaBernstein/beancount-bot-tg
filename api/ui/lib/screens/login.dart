import 'package:flutter/material.dart';
import '../models/constants.dart';
import '../service/api.dart';

class MyLoginPage extends StatefulWidget {
  const MyLoginPage({super.key /*, required this.title*/});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  /*final String title;*/

  @override
  State<MyLoginPage> createState() => _MyLoginState();
}

enum LoginButtonText {
  requestVerification(text: 'Request verification token'),
  submitVerification(text: 'Submit verification token and login');

  const LoginButtonText({
    required this.text,
  });

  final String text;
}

class _MyLoginState extends State<MyLoginPage> {
  String? userId;
  String? verificationCode;
  ClientAuthentication authentication = ClientAuthentication();
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  bool isVerificationCodeInputEnabled = false;
  String loginButtonText = LoginButtonText.requestVerification.text;

  void _updateToken() {
    setState(() {
      // This call to setState tells the Flutter framework that something has
      // changed in this State, which causes it to rerun the build method below
      // so that the display can reflect the updated values. If we changed
      // _counter without calling setState(), then the build method would not be
      // called again, and so nothing would appear to happen.
      if (authentication.token != null && authentication.token!.isNotEmpty) {
        _redirectHome();
      }
    });
  }

  void _redirectHome() {
    Navigator.pushNamedAndRemoveUntil(
        context, Routes.config.route, (Route<dynamic> route) => false);
  }

  void _checkExistingAuth() async {
    String? token = await ClientAuthentication.loadToken();
    if (token != null && token.isNotEmpty) {
      _redirectHome();
    }
  }

  void _switchLoginToValidation() {
    isVerificationCodeInputEnabled = true;
    loginButtonText = LoginButtonText.submitVerification.text;
  }

  @override
  Widget build(BuildContext context) {
    _checkExistingAuth();
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // TRY THIS: Try changing the color here to a specific color (to
        // Colors.amber, perhaps?) and trigger a hot reload to see the AppBar
        // change color while the other colors stay the same.
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text('Beancount-Bot-Tg Web-UI Login' /*widget.title*/),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: SizedBox(
          width: 500,
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
              const Text(
                'Authenticate to your Beancount-Bot-Tg instance:',
              ),
              Form(
                  key: _formKey,
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        TextFormField(
                          onSaved: (String? value) {
                            userId = value;
                          },
                          decoration: const InputDecoration(
                            hintText: 'Enter your user id',
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'user id is required';
                            }
                            return null;
                          },
                        ),
                        TextFormField(
                          onSaved: (String? value) {
                            verificationCode = value;
                          },
                          enabled: isVerificationCodeInputEnabled,
                          decoration: const InputDecoration(
                            hintText: 'Enter your verification code',
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 16.0),
                          child: ElevatedButton(
                            onPressed: () async {
                              // Validate will return true if the form is valid, or false if
                              // the form is invalid.
                              if (_formKey.currentState!.validate()) {
                                _formKey.currentState!.save();
                                String? error;
                                if (verificationCode != null &&
                                    verificationCode!.isNotEmpty) {
                                  (_, error) = await authentication
                                      .validateVerificationCode(
                                          verificationCode!);
                                } else {
                                  (error,) = await authentication
                                      .generateVerificationCode(userId!);
                                }
                                if (error != null && error.isNotEmpty) {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(content: Text(error)),
                                  );
                                } else if (error == null ||
                                    error.contains('(409)')) {
                                  // Challenge already in progress: Let user input verification code
                                  setState(() => _switchLoginToValidation());
                                }
                                _updateToken();
                              }
                            },
                            child: Text(loginButtonText),
                          ),
                        )
                      ])),
              const Text('Hint: In order to receive the credentials from your beancount-bot-tg instance, you need to activate the API first: For that, issue the following command:\n\n/config enable_api on'),
            ],
          ),
        ),
      ),
    );
  }
}
