import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  src: string;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Separation of concerns',
    Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
    // src: "https://www.svgrepo.com/download/520485/research.svg",
    // src: "https://www.svgrepo.com/download/514345/street-sign.svg",
    src: "/img/street-sign.svg",
    description: (
      <>
        DO was designed from the ground up to increase modularity
        by making things less connected and dependent.
      </>
    ),
  },
  {
    title: 'Application lifecycle',
    Svg: require('@site/static/img/undraw_docusaurus_tree.svg').default,
    // src: "https://www.svgrepo.com/download/520498/clock.svg",
    //src: "https://www.svgrepo.com/download/514347/compass.svg",
    src: "/img/compass.svg",
    description: (
      <>
        DO provides an API to check the health of active services and to
        unload an application gracefully in reverse dependency order.
      </>
    ),
  },
  {
    title: 'Easy debugging',
    Svg: require('@site/static/img/undraw_docusaurus_react.svg').default,
    // src: "https://www.svgrepo.com/download/520483/quiz.svg",
    // src: "https://www.svgrepo.com/download/514336/telescope.svg",
    src: "/img/telescope.svg",
    description: (
      <>
        Debugging IoC can be painful.
        DO offers an API to describe the application layout and visualize
        a service with its full dependency tree.
      </>
    ),
  },
];

function Feature({title, Svg,src, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <img src={src} className={styles.featureSvg} />
        {/* <Svg src="https://www.svgrepo.com/download/520496/board.svg" className={styles.featureSvg} role="img" /> */}
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
