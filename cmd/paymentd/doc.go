/*
   Copyright 2014 Fritz Payment GmbH

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/*
The paymentd daemon serves payment related services for the FritzPay stack.

Usage:
  paymentd

  Flags understood by paymentd:
    -c          Path to config file name.
                Alternatively the environment var $PAYMENTDCFG can be used to set
                the configuration file name.

  Example:
    paymentd -c /etc/paymentd/paymentd.config.json
*/
package main
