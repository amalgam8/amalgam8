//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package nginx

import (
	"github.com/amalgam8/amalgam8/pkg/api"
)

// MockClient mocks NGINX Client interface
type MockClient struct {
	UpdateError error
	UpdateCount int
}

// Update mocks interface
func (m *MockClient) Update([]api.ServiceInstance, []api.Rule) error {
	m.UpdateCount++
	return m.UpdateError
}
