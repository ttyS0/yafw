import Config from './pages/Config';
import Policy from './pages/Policy';
import Nat from './pages/Nat';
import Connection from './pages/Connection';

const routes = [
    {
        path: '/policy',
        title: 'Policies',
        component: <Policy />,
    },
    {
        path: '/snat',
        title: 'SNAT',
        component: <Nat />,
    },
    {
        path: '/config',
        title: 'Configuration',
        component: <Config />,
    },
    {
        path: '/connections',
        title: 'Connections',
        component: <Connection />,
    }
];

export default routes;